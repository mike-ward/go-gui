//go:build darwin && !ios

#import <AppKit/AppKit.h>
#include "menu_darwin.h"
#include <stdlib.h>
#include <string.h>

// Forward declaration of Go callback.
extern void goNativeMenuAction(const char *itemID);
extern void goNativeTrayAction(int trayID, const char *itemID);

// ---------------------------------------------------------------------------
// MenuActionHandler — singleton ObjC target for custom menu items.
// ---------------------------------------------------------------------------

@interface MenuActionHandler : NSObject
- (void)menuItemClicked:(id)sender;
@end

@implementation MenuActionHandler
- (void)menuItemClicked:(id)sender {
    NSMenuItem *item = (NSMenuItem *)sender;
    NSString *itemID = item.representedObject;
    if (itemID != nil) {
        goNativeMenuAction(itemID.UTF8String);
    }
}
@end

static MenuActionHandler *gHandler = nil;

static MenuActionHandler *sharedHandler(void) {
    if (gHandler == nil) {
        gHandler = [[MenuActionHandler alloc] init];
    }
    return gHandler;
}

// ---------------------------------------------------------------------------
// TrayActionHandler — per-tray ObjC target for tray menu items.
// ---------------------------------------------------------------------------

@interface TrayActionHandler : NSObject
@property (nonatomic, assign) int trayID;
- (void)menuItemClicked:(id)sender;
@end

@implementation TrayActionHandler
- (void)menuItemClicked:(id)sender {
    NSMenuItem *item = (NSMenuItem *)sender;
    NSString *itemID = item.representedObject;
    if (itemID != nil) {
        goNativeTrayAction(self.trayID, itemID.UTF8String);
    }
}
@end

// ---------------------------------------------------------------------------
// Helpers — build NSMenu from flat item array.
// ---------------------------------------------------------------------------

static NSString *shortcutString(char c) {
    if (c == 0) return @"";
    char buf[2] = {c, 0};
    // NSMenuItem key equivalents are lowercase.
    buf[0] = (char)tolower((unsigned char)c);
    return [NSString stringWithUTF8String:buf];
}

static NSEventModifierFlags shortcutModifiers(int mods) {
    NSEventModifierFlags flags = 0;
    if (mods & 1) flags |= NSEventModifierFlagCommand;
    if (mods & 2) flags |= NSEventModifierFlagShift;
    if (mods & 4) flags |= NSEventModifierFlagOption;
    if (mods & 8) flags |= NSEventModifierFlagControl;
    return flags;
}

// Recursively build an NSMenu from a slice of the flat array.
static NSMenu *buildMenu(NSString *title,
    NativeMenuItemC *items, int start, int count,
    NativeMenuItemC *allItems,
    id target, SEL action) {

    NSMenu *menu = [[NSMenu alloc] initWithTitle:title];
    menu.autoenablesItems = NO;

    for (int i = start; i < start + count; i++) {
        NativeMenuItemC *ci = &allItems[i];

        if (ci->separator) {
            [menu addItem:[NSMenuItem separatorItem]];
            continue;
        }

        NSString *text = ci->text
            ? [NSString stringWithUTF8String:ci->text]
            : @"";
        NSString *key = shortcutString(ci->shortcutChar);
        NSMenuItem *mi = [[NSMenuItem alloc]
            initWithTitle:text
            action:(ci->childCount > 0 ? nil : action)
            keyEquivalent:key];
        mi.keyEquivalentModifierMask =
            shortcutModifiers(ci->shortcutMods);
        mi.enabled = !ci->disabled;
        mi.state = ci->checked ? NSControlStateValueOn
                               : NSControlStateValueOff;
        mi.target = (ci->childCount > 0 ? nil : target);

        if (ci->id) {
            mi.representedObject =
                [NSString stringWithUTF8String:ci->id];
        }

        if (ci->childCount > 0) {
            NSMenu *sub = buildMenu(text,
                items, ci->childStart, ci->childCount,
                allItems, target, action);
            mi.submenu = sub;
        }

        [menu addItem:mi];
    }
    return menu;
}

// ---------------------------------------------------------------------------
// Standard Edit menu with responder-chain selectors.
// ---------------------------------------------------------------------------

static NSMenuItem *standardEditMenu(void) {
    NSMenu *edit = [[NSMenu alloc] initWithTitle:@"Edit"];

    NSMenuItem *undo = [[NSMenuItem alloc]
        initWithTitle:@"Undo" action:@selector(undo:)
        keyEquivalent:@"z"];
    undo.keyEquivalentModifierMask =
        NSEventModifierFlagCommand;
    [edit addItem:undo];

    NSMenuItem *redo = [[NSMenuItem alloc]
        initWithTitle:@"Redo" action:@selector(redo:)
        keyEquivalent:@"Z"];
    redo.keyEquivalentModifierMask =
        NSEventModifierFlagCommand | NSEventModifierFlagShift;
    [edit addItem:redo];

    [edit addItem:[NSMenuItem separatorItem]];

    NSMenuItem *cut = [[NSMenuItem alloc]
        initWithTitle:@"Cut" action:@selector(cut:)
        keyEquivalent:@"x"];
    cut.keyEquivalentModifierMask =
        NSEventModifierFlagCommand;
    [edit addItem:cut];

    NSMenuItem *copy_ = [[NSMenuItem alloc]
        initWithTitle:@"Copy" action:@selector(copy:)
        keyEquivalent:@"c"];
    copy_.keyEquivalentModifierMask =
        NSEventModifierFlagCommand;
    [edit addItem:copy_];

    NSMenuItem *paste = [[NSMenuItem alloc]
        initWithTitle:@"Paste" action:@selector(paste:)
        keyEquivalent:@"v"];
    paste.keyEquivalentModifierMask =
        NSEventModifierFlagCommand;
    [edit addItem:paste];

    [edit addItem:[NSMenuItem separatorItem]];

    NSMenuItem *selAll = [[NSMenuItem alloc]
        initWithTitle:@"Select All"
        action:@selector(selectAll:)
        keyEquivalent:@"a"];
    selAll.keyEquivalentModifierMask =
        NSEventModifierFlagCommand;
    [edit addItem:selAll];

    NSMenuItem *editItem = [[NSMenuItem alloc]
        initWithTitle:@"Edit" action:nil keyEquivalent:@""];
    editItem.submenu = edit;
    return editItem;
}

// ---------------------------------------------------------------------------
// App menu (About, Quit).
// ---------------------------------------------------------------------------

static NSMenuItem *appMenu(const char *appName) {
    NSString *name = appName
        ? [NSString stringWithUTF8String:appName]
        : @"";

    NSMenu *menu = [[NSMenu alloc] initWithTitle:name];

    NSString *aboutTitle =
        [NSString stringWithFormat:@"About %@", name];
    NSMenuItem *about = [[NSMenuItem alloc]
        initWithTitle:aboutTitle
        action:@selector(orderFrontStandardAboutPanel:)
        keyEquivalent:@""];
    [menu addItem:about];

    [menu addItem:[NSMenuItem separatorItem]];

    NSString *quitTitle =
        [NSString stringWithFormat:@"Quit %@", name];
    NSMenuItem *quit = [[NSMenuItem alloc]
        initWithTitle:quitTitle
        action:@selector(terminate:)
        keyEquivalent:@"q"];
    quit.keyEquivalentModifierMask =
        NSEventModifierFlagCommand;
    [menu addItem:quit];

    NSMenuItem *item = [[NSMenuItem alloc]
        initWithTitle:name action:nil keyEquivalent:@""];
    item.submenu = menu;
    return item;
}

// ---------------------------------------------------------------------------
// Public C API — menubar.
// ---------------------------------------------------------------------------

void nativemenuSetMenubar(const char *appName,
    NativeMenuItemC *menus, int menuCount,
    NativeMenuItemC *allItems, int itemCount,
    int includeEditMenu) {

    @autoreleasepool {
        NSMenu *bar = [[NSMenu alloc] init];

        // App menu.
        [bar addItem:appMenu(appName)];

        // User-defined menus.
        MenuActionHandler *handler = sharedHandler();
        SEL action = @selector(menuItemClicked:);
        for (int i = 0; i < menuCount; i++) {
            NativeMenuItemC *mc = &menus[i];
            NSString *title = mc->text
                ? [NSString stringWithUTF8String:mc->text]
                : @"";

            NSMenu *sub = buildMenu(title,
                allItems, mc->childStart, mc->childCount,
                allItems, handler, action);

            NSMenuItem *topItem = [[NSMenuItem alloc]
                initWithTitle:title action:nil
                keyEquivalent:@""];
            topItem.submenu = sub;
            [bar addItem:topItem];
        }

        // Standard Edit menu.
        if (includeEditMenu) {
            // Insert after File if present, or at index 1.
            NSInteger idx = bar.numberOfItems > 1 ? 2 : 1;
            [bar insertItem:standardEditMenu() atIndex:idx];
        }

        [NSApp setMainMenu:bar];
    }
}

void nativemenuClearMenubar(void) {
    @autoreleasepool {
        [NSApp setMainMenu:[[NSMenu alloc] init]];
    }
}

// ---------------------------------------------------------------------------
// System tray.
// ---------------------------------------------------------------------------

static NSMutableDictionary<NSNumber *, NSStatusItem *>
    *gTrayItems = nil;
static NSMutableDictionary<NSNumber *, TrayActionHandler *>
    *gTrayHandlers = nil;
static int gNextTrayID = 1;

static NSImage *imageFromPNG(const void *data, int len) {
    if (data == NULL || len <= 0) return nil;
    NSData *d = [NSData dataWithBytes:data length:len];
    NSImage *img = [[NSImage alloc] initWithData:d];
    if (img != nil) {
        // Template images adapt to menu bar appearance.
        [img setTemplate:YES];
        // Standard status item icon size.
        img.size = NSMakeSize(18, 18);
    }
    return img;
}

int nativemenuCreateTray(const void *iconData, int iconLen,
    const char *tooltip,
    NativeMenuItemC *items, int itemCount) {

    @autoreleasepool {
        if (gTrayItems == nil) {
            gTrayItems = [NSMutableDictionary dictionary];
            gTrayHandlers = [NSMutableDictionary dictionary];
        }

        int trayID = gNextTrayID++;
        NSNumber *key = @(trayID);

        NSStatusItem *si = [[NSStatusBar systemStatusBar]
            statusItemWithLength:NSVariableStatusItemLength];

        NSImage *img = imageFromPNG(iconData, iconLen);
        if (img != nil) si.button.image = img;

        if (tooltip != NULL) {
            si.button.toolTip =
                [NSString stringWithUTF8String:tooltip];
        }

        TrayActionHandler *handler =
            [[TrayActionHandler alloc] init];
        handler.trayID = trayID;

        if (items != NULL && itemCount > 0) {
            NSMenu *menu = buildMenu(@"",
                items, 0, itemCount, items,
                handler,
                @selector(menuItemClicked:));
            si.menu = menu;
        }

        gTrayItems[key] = si;
        gTrayHandlers[key] = handler;
        return trayID;
    }
}

void nativemenuUpdateTray(int trayID,
    const void *iconData, int iconLen,
    const char *tooltip,
    NativeMenuItemC *items, int itemCount) {

    @autoreleasepool {
        NSNumber *key = @(trayID);
        NSStatusItem *si = gTrayItems[key];
        if (si == nil) return;

        NSImage *img = imageFromPNG(iconData, iconLen);
        if (img != nil) si.button.image = img;

        if (tooltip != NULL) {
            si.button.toolTip =
                [NSString stringWithUTF8String:tooltip];
        }

        TrayActionHandler *handler = gTrayHandlers[key];
        if (items != NULL && itemCount > 0) {
            NSMenu *menu = buildMenu(@"",
                items, 0, itemCount, items,
                handler,
                @selector(menuItemClicked:));
            si.menu = menu;
        }
    }
}

void nativemenuRemoveTray(int trayID) {
    @autoreleasepool {
        NSNumber *key = @(trayID);
        NSStatusItem *si = gTrayItems[key];
        if (si == nil) return;
        [[NSStatusBar systemStatusBar] removeStatusItem:si];
        [gTrayItems removeObjectForKey:key];
        [gTrayHandlers removeObjectForKey:key];
    }
}

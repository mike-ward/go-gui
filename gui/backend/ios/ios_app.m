#import <UIKit/UIKit.h>
#import <Metal/Metal.h>
#import <QuartzCore/CAMetalLayer.h>
#include "ios_app.h"

// ─── GoGuiView ───────────────────────────────────────────────

@interface GoGuiView : UIView
@end

@implementation GoGuiView
+ (Class)layerClass { return [CAMetalLayer class]; }
@end

// ─── GoGuiViewController ─────────────────────────────────────

@interface GoGuiViewController : UIViewController
@property (nonatomic, strong) CADisplayLink *displayLink;
@property (nonatomic, assign) BOOL started;
@end

@implementation GoGuiViewController

- (void)loadView {
    GoGuiView *v = [[GoGuiView alloc] init];
    v.backgroundColor = [UIColor blackColor];
    v.multipleTouchEnabled = NO;
    self.view = v;
}

- (void)viewDidLayoutSubviews {
    [super viewDidLayoutSubviews];
    CGRect bounds = self.view.bounds;
    UIScreen *screen = self.view.window.screen;
    CGFloat scale = screen ? screen.scale : 2.0;

    CAMetalLayer *layer = (CAMetalLayer *)self.view.layer;
    layer.contentsScale = scale;
    layer.drawableSize = CGSizeMake(
        bounds.size.width * scale,
        bounds.size.height * scale);

    if (!self.started) {
        id<MTLDevice> device = MTLCreateSystemDefaultDevice();
        if (!device) return;

        layer.device = device;
        layer.pixelFormat = MTLPixelFormatBGRA8Unorm;
        layer.framebufferOnly = YES;

        void *layerPtr = (__bridge void *)layer;
        goIOSInit(layerPtr,
                  (int)bounds.size.width,
                  (int)bounds.size.height,
                  (float)scale);

        self.displayLink = [CADisplayLink
            displayLinkWithTarget:self
            selector:@selector(render:)];
        [self.displayLink addToRunLoop:[NSRunLoop mainRunLoop]
                               forMode:NSRunLoopCommonModes];

        [[NSNotificationCenter defaultCenter]
            addObserver:self
            selector:@selector(appWillResignActive:)
            name:UIApplicationWillResignActiveNotification
            object:nil];
        [[NSNotificationCenter defaultCenter]
            addObserver:self
            selector:@selector(appDidBecomeActive:)
            name:UIApplicationDidBecomeActiveNotification
            object:nil];

        self.started = YES;
    } else {
        goIOSResize((int)bounds.size.width,
                    (int)bounds.size.height,
                    (float)scale);
    }
}

- (void)render:(CADisplayLink *)link {
    goIOSRender();
}

- (void)touchesBegan:(NSSet<UITouch *> *)touches
           withEvent:(UIEvent *)event {
    UITouch *touch = [touches anyObject];
    if (!touch) return;
    CGPoint loc = [touch locationInView:self.view];
    goIOSTouchBegan((float)loc.x, (float)loc.y);
}

- (void)touchesMoved:(NSSet<UITouch *> *)touches
           withEvent:(UIEvent *)event {
    UITouch *touch = [touches anyObject];
    if (!touch) return;
    CGPoint loc = [touch locationInView:self.view];
    goIOSTouchMoved((float)loc.x, (float)loc.y);
}

- (void)touchesEnded:(NSSet<UITouch *> *)touches
           withEvent:(UIEvent *)event {
    UITouch *touch = [touches anyObject];
    if (!touch) return;
    CGPoint loc = [touch locationInView:self.view];
    goIOSTouchEnded((float)loc.x, (float)loc.y);
}

- (void)touchesCancelled:(NSSet<UITouch *> *)touches
               withEvent:(UIEvent *)event {
    UITouch *touch = [touches anyObject];
    if (!touch) return;
    CGPoint loc = [touch locationInView:self.view];
    goIOSTouchEnded((float)loc.x, (float)loc.y);
}

- (void)appWillResignActive:(NSNotification *)n {
    self.displayLink.paused = YES;
}

- (void)appDidBecomeActive:(NSNotification *)n {
    self.displayLink.paused = NO;
}

- (void)dealloc {
    [[NSNotificationCenter defaultCenter] removeObserver:self];
    [self.displayLink invalidate];
}

@end

// ─── GoGuiAppDelegate ────────────────────────────────────────

@interface GoGuiAppDelegate : UIResponder <UIApplicationDelegate>
@property (nonatomic, strong) UIWindow *window;
@end

@implementation GoGuiAppDelegate

- (BOOL)application:(UIApplication *)application
    didFinishLaunchingWithOptions:(NSDictionary *)opts {
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wdeprecated-declarations"
    self.window = [[UIWindow alloc]
        initWithFrame:[UIScreen mainScreen].bounds];
#pragma clang diagnostic pop
    self.window.rootViewController =
        [[GoGuiViewController alloc] init];
    [self.window makeKeyAndVisible];
    return YES;
}

@end

// ─── Entry Point ─────────────────────────────────────────────

void iosStartApp(void) {
    @autoreleasepool {
        char *argv[] = {""};
        UIApplicationMain(1, argv, nil,
            NSStringFromClass([GoGuiAppDelegate class]));
    }
}

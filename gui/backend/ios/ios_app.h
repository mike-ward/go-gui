#ifndef IOS_APP_H
#define IOS_APP_H

// Start the UIKit application. Calls UIApplicationMain which
// blocks forever. The UIKit lifecycle calls back into Go via
// the goIOS* exported functions.
void iosStartApp(void);

// Exported Go callbacks — implemented in backend.go.
extern void goIOSInit(void* layer, int w, int h, float scale);
extern void goIOSRender(void);
extern void goIOSResize(int w, int h, float scale);
extern void goIOSTouchBegan(float x, float y);
extern void goIOSTouchMoved(float x, float y);
extern void goIOSTouchEnded(float x, float y);

#endif

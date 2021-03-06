#include "gui.h"

#import <Cocoa/Cocoa.h>
#import "window.h"
#import "windowDelegate.h"
#import "glView.h"
#import "textView.h"
#import "blurView.h"
#import "progressView.h"
#import "volumeView.h"
// #import "titleTextView.h"
#import "app.h"

void initialize() {
        if (NSApp)
                return;

        [NSApplication sharedApplication];

        [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];

        Application *appDelegate = [[Application alloc] init];
        [NSApp setDelegate:appDelegate];

        [[NSBundle mainBundle] loadNibNamed:@"MainMenu" owner:NSApp topLevelObjects:nil];

        [[NSAppleEventManager sharedAppleEventManager] setEventHandler:appDelegate andSelector:@selector(handleAppleEvent:withReplyEvent:) forEventClass:kInternetEventClass andEventID:kAEGetURL];
}

void setWindowTitleWithRepresentedFilename(void* wptr, char* title) {
        Window* w = (Window*)wptr;

        NSString* str = [NSString stringWithUTF8String:title];
	[w setTitleWithRepresentedFilename:str];
}

void setWindowTitle(void* wptr, char* title) {
        Window* w = (Window*)wptr;

        NSString* str = [NSString stringWithUTF8String:title];
        [w setTitle:str];
}

void setWindowSize(void* wptr, int width, int height) {
        Window* w = (Window*)wptr;

        NSRect frame = [w frame];

        frame.origin.y -= (height - frame.size.height)/2;
        frame.origin.x -= (width - frame.size.width)/2;

        if (frame.origin.x < 50) {
                frame.origin.x = 50;
        }

        CGFloat screenh = [[NSScreen mainScreen] frame].size.height;

        if (frame.origin.y > screenh - height - 80) {
                frame.origin.y = screenh - height - 80;
        }

        frame.size = NSMakeSize(width, height);

        w.aspectRatio = NSMakeSize(width, height);
        [w->glView setOriginalSize:NSMakeSize(width, height)];

        [w setFrame:frame display:YES animate:YES];

        [w makeKeyAndOrderFront:nil];
}

void* newWindow(char* title, int width, int height) {
        @autoreleasepool {

                initialize();

                Window* w = [[Window alloc] initWithWidth:width height:height];
                setWindowTitle(w, title);

                WindowDelegate* wd = (WindowDelegate*)[[WindowDelegate alloc] init];
                [w setDelegate:(id)wd];

                [w makeFirstResponder:w->glView];

                NSTimer *renderTimer = [NSTimer timerWithTimeInterval:1.0/30.0 
                                                               target:w->glView
                                                             selector:@selector(timerTick:)
                                                             userInfo:nil
                                                              repeats:YES];

                [[NSRunLoop currentRunLoop] addTimer:renderTimer
                                             forMode:NSDefaultRunLoopMode];
                [[NSRunLoop currentRunLoop] addTimer:renderTimer
                                             forMode:NSEventTrackingRunLoopMode]; //Ensure timer fires during resize

                return w;
        }
}

void showWindow(void* ptr) {
        [NSApp activateIgnoringOtherApps:YES];

        Window* w = (Window*)ptr;
        [w makeKeyAndOrderFront:nil];
}
void initWindowCurrentContext(void*ptr) {
        Window* w = (Window*)ptr;
        [w makeCurrentContext];
}
void makeCurrentContext(void* ptr) {
        Window* w = (Window*)ptr;
        [w->glView makeCurrentContext];
}
void flushBuffer(void* ptr) {
        Window* w = (Window*)ptr;
        [w->glView flushBuffer];
} 
void pollEvents() {
        [NSApp run];
        // NSApplicationMain(0, NULL);
}
void refreshWindowContent(void*wptr) {
        Window* w = (Window*)wptr;
        [w->glView setNeedsDisplay:YES];
}

CSize getWindowSize(void* ptr) {
        Window* w = (Window*)ptr;
        CSize sz;
        sz.width = (int)([w->glView frame].size.width);
        sz.height = (int)([w->glView frame].size.height);
        return sz;
}

void updatePlaybackInfo(void* ptr, char* left, char* right, double percent) {
        Window* w = (Window*)ptr;

        NSString* leftStr;
        if (strlen(left) == 0) {
                leftStr = @"00:00:00";
        } else {
                leftStr = [[NSString stringWithUTF8String:left] retain];
        }
        NSString* rightStr;
        if (strlen(right) == 0) {
                rightStr = @"00:00:00";
        } else {
                rightStr = [[NSString stringWithUTF8String:right] retain];
        }
        [w->glView updatePorgressInfo:leftStr rightString:rightStr percent:percent];
}
void updateBufferInfo(void* ptr, char* speed, double percent) {
        Window* w = (Window*)ptr;
        NSString* str;
        if (strlen(speed) == 0) {
                str = @"";
        } else {
                str = [[NSString stringWithUTF8String:speed] retain];
        }
        [w->glView updateBufferInfo:str bufferPercent:percent];
}
void* showSubtitle(void* ptr, SubItem* item) {
        Window* w = (Window*)ptr;
        return [w->glView showSubtitle:item];
}
void hideSubtitle(void* ptrWin, long ptrText) {
        Window* w = (Window*)ptrWin;
        [w->glView hideSubtitle:(TextView*)ptrText];
}
void setSpinningVisible(void* ptr, int b) {
        Window* w = (Window*)ptr;
        [w->glView setSpinningHidden:(b==0)];
}
void toggleFullScreen(void* ptr) {
        Window* w = (Window*)ptr;
        [w toggleFullScreen:nil];
}

int isFullScreen(void* ptr) {
        Window* w = (Window*)ptr;
        return (int)[w isFullScreen];
}

void setControlsVisible(void* ptr, int b, int autoHide) {    
        Window* w = (Window*)ptr;

        BOOL hidden = (b==0);

        if (!hidden) {
                if (autoHide) {
                        [w->glView setShowCursorDeadline:[NSDate dateWithTimeIntervalSinceNow:(NSTimeInterval)2.0]];
                } else {
                        [w->glView setShowCursorDeadline:[NSDate distantFuture]];
                        NSLog(@"never hide automatically");
                }
        }

        if ((hidden && [w->glView isCursorHidden]) || (!hidden && ![w->glView isCursorHidden])) {
                return;
        }

        [w->glView setCursorHidden:hidden];
        [w->glView setPlaybackViewHidden:hidden];

        if (![w isFullScreen])
                [w setTitleHidden:hidden];
}

CSize getScreenSize() {
        NSSize sz = [[NSScreen mainScreen] frame].size;
        CSize csz;
        csz.width = (int)sz.width;
        csz.height = (int)sz.height;
        return csz;
}

void setVolume(void* wptr, int volume) {
        Window* w = (Window*)wptr;
        [w->glView setVolume:volume];
}

void setVolumeVisible(void* wptr, int b) {
        Window* w = (Window*)wptr;
        [w->glView setVolumeHidden:(b==0)];
}

void alert(void* wptr, char* str) {
        Window* w = (Window*)wptr;

        setControlsVisible(w, 1, 0);
        [w fatal:[NSString stringWithUTF8String:str]];
}

void closeWindow(void* wptr) {
        Window* w = (Window*)wptr;
        [w close];
}

void addRecentOpenedFile(char* str) {
        NSString* filename = [NSString stringWithUTF8String:str];
        [[NSDocumentController sharedDocumentController] noteNewRecentDocumentURL:[NSURL fileURLWithPath:filename]];
}

void setSubFontSize(void* wptr, double sz) {
        if (sz == 0) sz = 25;

        Window* w = (Window*)wptr;
        [w->glView setFontSize:sz];
}

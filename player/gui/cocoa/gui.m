#include "gui.h"

#import <Cocoa/Cocoa.h>
#import "window.h"
#import "windowDelegate.h"
#import "glView.h"
#import "textView.h"
#import "blurView.h"
#import "progressView.h"
#import "startupView.h"
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

	//create memu bar
	id menubar = [[NSMenu new] autorelease];
    id appMenuItem = [[NSMenuItem new] autorelease];
    [menubar addItem:appMenuItem];
    [NSApp setMainMenu:menubar];
    id appMenu = [[NSMenu new] autorelease];

    NSMenuItem *searchSubtitleMenuItem = [[[NSMenuItem alloc] initWithTitle:@"Search Subtitle"
        action:@selector(searchSubtitleMenuItemClick:) keyEquivalent:@"s"] autorelease];
    [searchSubtitleMenuItem setKeyEquivalentModifierMask:0];
    [searchSubtitleMenuItem setTarget: appDelegate];
    [appMenu addItem:searchSubtitleMenuItem];

    NSMenuItem *openFileMenuItem = [[[NSMenuItem alloc] initWithTitle:@"Open..."
        action:@selector(openFileMenuItemClick:) keyEquivalent:@"o"] autorelease];
    [openFileMenuItem setTarget: appDelegate];
    [appMenu addItem:openFileMenuItem];

    id appName = [[NSProcessInfo processInfo] processName];
    id quitTitle = [@"Quit " stringByAppendingString:appName];
    id quitMenuItem = [[[NSMenuItem alloc] initWithTitle:quitTitle
        action:@selector(terminate:) keyEquivalent:@"q"] autorelease];
    [appMenu addItem:quitMenuItem];


    [appMenuItem setSubmenu:appMenu];

    [[NSAppleEventManager sharedAppleEventManager] setEventHandler:appDelegate andSelector:@selector(handleAppleEvent:withReplyEvent:) forEventClass:kInternetEventClass andEventID:kAEGetURL];
}
NSMenuItem* getTopMenuByTitle(NSString* title) {
    NSMenu* menubar = [NSApp mainMenu];
    NSArray* menus = [menubar itemArray];
    for (NSMenuItem* menu in menus) {
        if ([menu title] == title) {
            return menu;
        }
    }

    return nil;
}
void hideMenuNSString(NSString* title) {
    NSAutoreleasePool * pool = [[NSAutoreleasePool alloc] init];
    // NSMenu* menubar = [NSApp mainMenu];
    // NSArray* menus = [menubar itemArray];
    NSMenuItem* item = getTopMenuByTitle(title);
    if (item != nil) {
        [[NSApp mainMenu] removeItem:item];
    }
    // for (NSMenuItem* menu in menus) {
    //     NSLog(@"compare %@ to %@", [menu title], title);
    //     if ([menu title] == title) {
    //         NSLog(@"remove menu item");
    //         [menubar removeItem:menu];
    //         break;
    //     }
    // }
    [pool drain];
}
void hideSubtitleMenu() {
    hideMenuNSString(@"Subtitle");
}
void hideAudioMenu() {
    hideMenuNSString(@"Audio");
}
void initAudioMenu(void* wptr, char** names, int32_t* tags, int len, int selected) {
    hideAudioMenu();

    NSAutoreleasePool * pool = [[NSAutoreleasePool alloc] init];
    if (len > 0) {
        NSWindow* w = (NSWindow*)wptr;

        NSMenu *menubar = [NSApp mainMenu];
        NSMenuItem* audioMenuItem = [[NSMenuItem new] autorelease];
        [audioMenuItem setTitle:@"Audio"];
        [menubar addItem:audioMenuItem];
        NSMenu* audioMenu = [[NSMenu alloc] initWithTitle:@"Audio"];

        for (int i = 0; i < len; i++) {
            char* name = names[i];
            int tag = tags[i];
            NSMenuItem* item = [[NSMenuItem alloc] initWithTitle:[NSString stringWithUTF8String:name] 
                action:@selector(audioMenuItemClick:) keyEquivalent:@""];
            [item setTarget: w];
            [item setTag: tag];
            [audioMenu addItem:item];

            if (tag == selected) {
                [item setState: NSOnState];
            }
        }
        [audioMenuItem setSubmenu:audioMenu];
    }
    [pool drain];
}

void initSubtitleMenu(void* wptr, char** names, int32_t* tags, int len, int32_t selected1, int32_t selected2) {
    hideSubtitleMenu();
    
    NSAutoreleasePool * pool = [[NSAutoreleasePool alloc] init];
    if (len > 0) {
        NSWindow* w = (NSWindow*)wptr;

        NSMenu* menubar = [NSApp mainMenu];
        NSMenuItem* subtitleMenuItem = [[NSMenuItem new] autorelease];
        [subtitleMenuItem setTitle:@"Subtitle"];
        [menubar addItem:subtitleMenuItem];
        NSMenu* subtitleMenu = [[NSMenu alloc] initWithTitle:@"Subtitle"];

        for (int i = 0; i < len; i++) {
            char* name = names[i];
            int tag = tags[i];
            NSMenuItem* item = [[NSMenuItem alloc] initWithTitle:[NSString stringWithUTF8String:name] 
                action:@selector(subtitleMenuItemClick:) keyEquivalent:@""];
            [item autorelease];
            [item setTarget: w];
            [item setTag: tag];
            [subtitleMenu addItem:item];

            if (tag == selected1 || tag == selected2) {
                [item setState: NSOnState];
            }
        }

        [subtitleMenuItem setSubmenu:subtitleMenu];
    }

    [pool drain];
}
void setSubtitleMenuItem(int t1, int t2) {
    NSAutoreleasePool * pool = [[NSAutoreleasePool alloc] init];

    NSMenuItem* menu = getTopMenuByTitle(@"Subtitle");
    for (NSMenuItem* item in [[menu submenu] itemArray]) {
        int tag = (int)[item tag];
        if (tag == t1 || tag == t2) {
            [item setState:NSOnState];
        } else {
            [item setState:NSOffState];
        }
    }
    
    [pool drain];
}

void setWindowTitle(void* wptr, char* title) {
    Window* w = (Window*)wptr;

    NSString* str = [NSString stringWithUTF8String:title];
    w->glView->titleTextView.title = str;
}

void setWindowSize(void* wptr, int width, int height) {
    Window* w = (Window*)wptr;

    NSRect frame = [w frame];
    frame.origin.y -= (height - frame.size.height)/2;
    frame.origin.x -= (width - frame.size.width)/2;
    frame.size = NSMakeSize(width, height);

    w->customAspectRatio = NSMakeSize(width, height);
    [w->glView setOriginalSize:NSMakeSize(width, height)];

    [w setFrame:frame display:YES animate:YES];

    [w makeKeyAndOrderFront:nil];
}

void* newWindow(char* title, int width, int height) {
	NSAutoreleasePool * pool = [[NSAutoreleasePool alloc] init];

	initialize();

	Window* w = [[Window alloc] initWithTitle:[NSString stringWithUTF8String:title]
		width:width height:height];
	
    WindowDelegate* wd = (WindowDelegate*)[[WindowDelegate alloc] init];
	[w setDelegate:(id)wd];

	GLView* v = [[GLView alloc] initWithFrame2:NSMakeRect(0,0,width,height)];
    w->glView = v;
	// [w setContentView:v];
    // BlurView* topbv = [[BlurView alloc] initWithFrame:NSMakeRect(0, height-30,width,30)];
    // w->titlebarView = topbv;

    NSView* rv = [[w contentView] superview];

    v->frameView = rv;

    NSView* roundView = [[NSView alloc] initWithFrame:NSMakeRect(0,0,width,height)];
    roundView.wantsLayer = YES;
    roundView.layer.masksToBounds = YES;
    roundView.layer.cornerRadius = 4.1;
    [roundView addSubview:v];

    [rv addSubview:roundView positioned:NSWindowBelow relativeTo:nil];

    // [rv addSubview:topbv];


    [roundView setFrame:NSMakeRect(0, 0, width, height)];
    [roundView setAutoresizingMask:NSViewWidthSizable|NSViewHeightSizable];
    [v setFrame:NSMakeRect(0,0,width,height)];
    [v setAutoresizingMask:NSViewWidthSizable|NSViewHeightSizable];


    [rv setWantsLayer:YES];
    rv.layer.cornerRadius=4.1;
    rv.layer.masksToBounds=YES;

    BlurView* bv = [[BlurView alloc] initWithFrame:NSMakeRect(0,0,width,22)];
    // bv.tintColor = [NSColor whiteColor];
    [bv setAutoresizingMask:NSViewWidthSizable|NSViewMaxYMargin];
    ProgressView* pv = [[ProgressView alloc] initWithFrame:NSMakeRect(0,0,width,bv.frame.size.height)];
    [bv addSubview:pv positioned:NSWindowBelow relativeTo:nil];
    [pv setAutoresizingMask:NSViewWidthSizable|NSViewHeightSizable];
    [v addSubview:bv positioned:NSWindowAbove relativeTo:nil];
    [v setProgressView:pv];
    v->blurView = bv;

    BlurView* tiv = [[BlurView alloc] initWithFrame:NSMakeRect(0,height-22,width,22)];
    [tiv setAutoresizingMask:NSViewWidthSizable|NSViewMinYMargin];
    // tiv.tintColor = [NSColor whiteColor];
    // [tiv setAutoresizingMask:NSViewMinXMargin | NSViewMaxXMargin | NSViewMinYMargin];
    TitleTextView* ttv = [[TitleTextView alloc] initWithFrame:NSMakeRect(0,0,width,22)];
    [tiv addSubview:ttv positioned:NSWindowBelow relativeTo:nil];
    [ttv setAutoresizingMask:NSViewWidthSizable|NSViewHeightSizable];
    // [v addSubview:tiv positioned:NSWindowAbove relativeTo:nil];
    // [v addSubview:tiv positioned:NSWindowAbove relativeTo:nil];
    [v addSubview:tiv positioned:NSWindowAbove relativeTo:nil];
    ttv.title = @"V'ger";
    v->titleTextView = ttv;
    v->titleView = tiv;

    SpinningView* spv = [[SpinningView alloc] initWithFrame:NSMakeRect((width-50)/2, (height-50)/2, 50, 50)];
    [spv setAutoresizingMask:NSViewMinXMargin|NSViewMaxXMargin|NSViewMinYMargin|NSViewMaxYMargin];
    [v addSubview:spv positioned:NSWindowAbove relativeTo:nil];
    v->spinningView = spv;
    [spv setHidden:YES];

    BlurView* bv2 = [[BlurView alloc] initWithFrame:NSMakeRect((width-120)/2, (height-120)/2, 120, 120)];
    [bv2 setBlurRadius:30.0];
    [bv2 setAutoresizingMask:NSViewMinXMargin|NSViewMaxXMargin|NSViewMinYMargin|NSViewMaxYMargin];    
    bv2.wantsLayer = YES;
    bv2.layer.masksToBounds = YES;
    bv2.layer.cornerRadius = 4.1;
    VolumeView* vv = [[VolumeView alloc] initWithFrame:NSMakeRect(0, 0, 120, 120)];
    [vv setAutoresizingMask:NSViewWidthSizable|NSViewHeightSizable];
    [v addSubview:bv2 positioned:NSWindowAbove relativeTo:spv];
    [bv2 addSubview:vv];
    v->volumeView = vv;
    v->volumeView2 = bv2;
    [bv2 setHidden:YES];

    [w makeFirstResponder:v];
    v->win = w;

    // [v addSubview:bvPopup];
    // [bvPopup setAutoresizingMask:NSViewWidthSizable];

    // PopupView* ppv = [[PopupView alloc] initWithFrame:NSMakeRect(0,0,400,500)];
    // [bvPopup addSubview:ppv];
    // [ppv setAutoresizingMask:NSViewWidthSizable|NSViewHeightSizable];


    StartupView* sv = [[StartupView alloc] initWithFrame:[v frame]];
    [v addSubview:sv positioned:NSWindowAbove relativeTo:bv];
    [sv setAutoresizingMask:NSViewWidthSizable|NSViewHeightSizable];

    [v setStartupView:sv];
    // [sv setNeedsDisplay:NO];

    [w setTitle:@""];

    NSTimer *renderTimer = [NSTimer timerWithTimeInterval:1.0/100.0 
                            target:w
                          selector:@selector(timerTick:)
                          userInfo:nil
                           repeats:YES];

    [[NSRunLoop currentRunLoop] addTimer:renderTimer
                                forMode:NSDefaultRunLoopMode];
    [[NSRunLoop currentRunLoop] addTimer:renderTimer
                                forMode:NSEventTrackingRunLoopMode]; //Ensure timer fires during resize


	[pool drain];

	return w;
}

void showWindow(void* ptr) {
	[NSApp activateIgnoringOtherApps:YES];

	Window* w = (Window*)ptr;
	[w makeKeyAndOrderFront:nil];
}
void makeWindowCurrentContext(void*ptr) {
    Window* w = (Window*)ptr;
    [w makeCurrentContext];
}
void pollEvents() {
    NSAutoreleasePool* pool = [[NSAutoreleasePool alloc] init];
    [NSApp finishLaunching];

    while(YES) {
	    [pool drain];
		pool = [[NSAutoreleasePool alloc] init];

	    NSEvent* event = [NSApp nextEventMatchingMask:NSAnyEventMask
	                                        untilDate:[NSDate distantFuture]
	                                        inMode:NSDefaultRunLoopMode
	                                        dequeue:YES];
	    [NSApp sendEvent:event];
	}
	// [NSApp activateIgnoringOtherApps:YES];
    //[NSApp run];

    [pool drain];
}
void refreshWindowContent(void*wptr) {
	Window* w = (Window*)wptr;
	[w setContentViewNeedsDisplay:YES];
}

int getWindowWidth(void* ptr) {
    Window* w = (Window*)ptr;
    return (int)([w->glView frame].size.width);
}
int getWindowHeight(void* ptr) {
    Window* w = (Window*)ptr;
    return (int)([w->glView frame].size.height);
}
void showWindowProgress(void* ptr, char* left, char* right, double percent) {
    Window* w = (Window*)ptr;
    [w->glView showProgress:left right:right percent:percent];
}
void showWindowBufferInfo(void* ptr, char* speed, double percent) {
    Window* w = (Window*)ptr;
    [w->glView showBufferInfo:speed bufferPercent:percent];
}
void* showText(void* ptr, SubItem* item) {
    Window* w = (Window*)ptr;
    return [w->glView showText:item];
}
void hideText(void* ptrWin, void* ptrText) {
    Window* w = (Window*)ptrWin;
    [w->glView hideText:ptrText];
}
void windowHideStartupView(void* ptr) {
    Window* w = (Window*)ptr;
    [w->glView hideStartupView];
}
void windowShowStartupView(void* ptr) {
    Window* w = (Window*)ptr;
    [w->glView showStartupView];
}
void showSpinning(void* ptr) {
    Window* w = (Window*)ptr;
    [w->glView->spinningView setHidden:NO];
}
void hideSpinning(void* ptr) {
    Window* w = (Window*)ptr;
    [w->glView->spinningView setHidden:YES];
}
void windowToggleFullScreen(void* ptr) {
    Window* w = (Window*)ptr;
    [w toggleFullScreen:nil];
}

void hideCursor(void* ptr) {
    Window* w = (Window*)ptr;
    [w->glView hideCursor];
    [w->glView hideProgress];
}

void showCursor(void* ptr) {
    Window* w = (Window*)ptr;
    [w->glView showCursor];
    [w->glView showProgress];
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

    w->glView->volumeView->_volume = volume;
    [w->glView->volumeView setNeedsDisplay:YES];
}

void setVolumeDisplay(void* wptr, int show) {
    Window* w = (Window*)wptr;
    BlurView* bv2 = w->glView->volumeView2;

    if (show != 0) {
        [bv2 setHidden:NO];
        NSSize sz = w.frame.size;
        [bv2 setFrame:NSMakeRect((sz.width-120)/2, (sz.height-120)/2, 120, 120)];
        [w->glView->volumeView setNeedsDisplay:YES];
    } else {
        [bv2 setHidden:YES];
    }
}

void alert(void* wptr, char* str) {
    Window* w = (Window*)wptr;
    [w setDelegate:nil];  //remove delegate prevent hide title bar
    [w->glView showTitle];

    NSAlert* alert = [[NSAlert alloc] init];
    [alert setMessageText:[NSString stringWithUTF8String:str]];
    [alert setAlertStyle:NSCriticalAlertStyle];

    [alert beginSheetModalForWindow:w modalDelegate:w didEndSelector:@selector(close) contextInfo:nil];
}

void closeWindow(void* wptr) {
    Window* w = (Window*)wptr;
    [w close];
}
//go:build darwin

package gui

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework UniformTypeIdentifiers

#import <Cocoa/Cocoa.h>
#import <UniformTypeIdentifiers/UniformTypeIdentifiers.h>

// Forward declarations for Go exports (defined in dockmenu_callbacks_darwin.go)
extern int goGetWindowCount(void);
extern const char* goGetWindowID(int index);
extern const char* goGetWindowTitle(int index);
extern void goDockMenuActivate(const char* windowID);
extern void goDockMenuClose(const char* windowID);
extern void goDockMenuOpenFile(const char* path);
extern void goDockMenuQuit(void);

// HostDelegate replaces AccessoryDelegate for multi-window mode.
@interface HostDelegate : NSObject <NSApplicationDelegate>
@end

@implementation HostDelegate

- (BOOL)applicationShouldTerminateAfterLastWindowClosed:(NSApplication *)sender {
    return NO;
}

- (NSMenu *)applicationDockMenu:(NSApplication *)sender {
    NSMenu *menu = [[NSMenu alloc] init];
    int count = goGetWindowCount();

    // Window list section
    for (int i = 0; i < count; i++) {
        const char *cTitle = goGetWindowTitle(i);
        const char *cID = goGetWindowID(i);
        if (!cTitle || !cID) continue;

        NSString *title = [NSString stringWithUTF8String:cTitle];
        NSString *wid = [NSString stringWithUTF8String:cID];

        NSMenuItem *item = [[NSMenuItem alloc]
            initWithTitle:title
            action:@selector(activateWindowAction:)
            keyEquivalent:@""];
        item.target = self;
        item.representedObject = wid;
        [menu addItem:item];

        free((void *)cTitle);
        free((void *)cID);
    }

    if (count > 0) {
        [menu addItem:[NSMenuItem separatorItem]];
    }

    // Close window section
    for (int i = 0; i < count; i++) {
        const char *cTitle = goGetWindowTitle(i);
        const char *cID = goGetWindowID(i);
        if (!cTitle || !cID) continue;

        NSString *title = [NSString stringWithFormat:@"Close %@",
            [NSString stringWithUTF8String:cTitle]];
        NSString *wid = [NSString stringWithUTF8String:cID];

        NSMenuItem *item = [[NSMenuItem alloc]
            initWithTitle:title
            action:@selector(closeWindowAction:)
            keyEquivalent:@""];
        item.target = self;
        item.representedObject = wid;
        [menu addItem:item];

        free((void *)cTitle);
        free((void *)cID);
    }

    if (count > 0) {
        [menu addItem:[NSMenuItem separatorItem]];
    }

    // Open File...
    NSMenuItem *openItem = [[NSMenuItem alloc]
        initWithTitle:@"Open File\u2026"
        action:@selector(openFileAction:)
        keyEquivalent:@""];
    openItem.target = self;
    [menu addItem:openItem];

    // Quit
    NSMenuItem *quitItem = [[NSMenuItem alloc]
        initWithTitle:@"Quit"
        action:@selector(quitAction:)
        keyEquivalent:@""];
    quitItem.target = self;
    [menu addItem:quitItem];

    return menu;
}

- (void)activateWindowAction:(NSMenuItem *)sender {
    NSString *wid = sender.representedObject;
    if (wid) {
        goDockMenuActivate([wid UTF8String]);
    }
}

- (void)closeWindowAction:(NSMenuItem *)sender {
    NSString *wid = sender.representedObject;
    if (wid) {
        goDockMenuClose([wid UTF8String]);
    }
}

- (void)openFileAction:(NSMenuItem *)sender {
    NSOpenPanel *panel = [NSOpenPanel openPanel];
    panel.allowsMultipleSelection = NO;
    panel.canChooseDirectories = NO;
    panel.canChooseFiles = YES;
    panel.allowedContentTypes = @[
        [UTType typeWithFilenameExtension:@"md"],
        [UTType typeWithFilenameExtension:@"markdown"],
    ];

    [panel beginWithCompletionHandler:^(NSModalResponse result) {
        if (result == NSModalResponseOK && panel.URL) {
            const char *path = [panel.URL.path UTF8String];
            goDockMenuOpenFile(path);
        }
    }];
}

- (void)quitAction:(NSMenuItem *)sender {
    goDockMenuQuit();
}

@end

static void guiInitHostMode(void) {
    [NSApplication sharedApplication];
    [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];
    [NSApp setDelegate:[[HostDelegate alloc] init]];
    [NSApp activateIgnoringOtherApps:YES];
}
*/
import "C"

func initHostMode() {
	C.guiInitHostMode()
	setDockIcon()
}

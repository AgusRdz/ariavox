//go:build darwin

package srbus

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit

#import <AppKit/AppKit.h>

void ariavox_announce(const char *text) {
    @autoreleasepool {
        NSString *msg = [NSString stringWithUTF8String:text];
        NSDictionary *info = @{
            NSAccessibilityAnnouncementKey: msg,
            NSAccessibilityPriorityKey: @(NSAccessibilityPriorityHigh)
        };
        NSAccessibilityPostNotificationWithUserInfo(
            [NSApp mainWindow] ?: [NSNull null],
            NSAccessibilityAnnouncementRequestedNotification,
            info
        );
    }
}
*/
import "C"

import "unsafe"

// NSAccessibilityBus posts VoiceOver announcements via NSAccessibility on macOS.
type NSAccessibilityBus struct{}

// New returns the platform-appropriate Bus.
// On macOS this always succeeds — VoiceOver may or may not be running,
// but the API call is always safe.
func New() Bus {
	return &NSAccessibilityBus{}
}

// Announce posts an NSAccessibilityAnnouncementRequestedNotification.
// VoiceOver intercepts this and speaks the text immediately if it is running.
// If VoiceOver is off this is a silent no-op.
func (b *NSAccessibilityBus) Announce(text string) error {
	if text == "" {
		return nil
	}
	cs := C.CString(text)
	defer C.free(unsafe.Pointer(cs))
	C.ariavox_announce(cs)
	return nil
}

func (b *NSAccessibilityBus) Close() error { return nil }

// Ensure NSAccessibilityBus satisfies Bus at compile time.
var _ Bus = (*NSAccessibilityBus)(nil)

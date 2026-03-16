//go:build darwin

package gui

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework CoreGraphics -framework CoreText

#import <Cocoa/Cocoa.h>
#import <CoreGraphics/CoreGraphics.h>
#import <CoreText/CoreText.h>

static void guiSetDockIcon(void) {
    int size = 512;
    CGColorSpaceRef space = CGColorSpaceCreateDeviceRGB();
    CGContextRef ctx = CGBitmapContextCreate(NULL, size, size, 8, size * 4, space,
        (CGBitmapInfo)kCGImageAlphaPremultipliedLast);
    CGColorSpaceRelease(space);
    if (!ctx) return;

    // Background: rounded rectangle with dark gradient
    CGFloat radius = size * 0.22;
    CGRect rect = CGRectMake(0, 0, size, size);
    CGMutablePathRef path = CGPathCreateMutable();
    CGPathMoveToPoint(path, NULL, radius, 0);
    CGPathAddLineToPoint(path, NULL, size - radius, 0);
    CGPathAddArc(path, NULL, size - radius, radius, radius, -M_PI_2, 0, false);
    CGPathAddLineToPoint(path, NULL, size, size - radius);
    CGPathAddArc(path, NULL, size - radius, size - radius, radius, 0, M_PI_2, false);
    CGPathAddLineToPoint(path, NULL, radius, size);
    CGPathAddArc(path, NULL, radius, size - radius, radius, M_PI_2, M_PI, false);
    CGPathAddLineToPoint(path, NULL, 0, radius);
    CGPathAddArc(path, NULL, radius, radius, radius, M_PI, M_PI + M_PI_2, false);
    CGPathCloseSubpath(path);

    CGContextSaveGState(ctx);
    CGContextAddPath(ctx, path);
    CGContextClip(ctx);

    // Gradient: dark charcoal top to slightly lighter bottom
    CGFloat colors[] = {
        0.14, 0.15, 0.17, 1.0,  // top: dark charcoal
        0.20, 0.21, 0.24, 1.0,  // bottom: slightly lighter
    };
    CGColorSpaceRef gradSpace = CGColorSpaceCreateDeviceRGB();
    CGGradientRef gradient = CGGradientCreateWithColorComponents(gradSpace, colors, NULL, 2);
    CGContextDrawLinearGradient(ctx, gradient, CGPointMake(0, size), CGPointMake(0, 0), 0);
    CGGradientRelease(gradient);
    CGColorSpaceRelease(gradSpace);
    CGContextRestoreGState(ctx);

    // "MD" text — bold system font, white, upper portion
    CTFontRef mdFont = CTFontCreateWithName(CFSTR("HelveticaNeue-Bold"), size * 0.32, NULL);
    NSDictionary *mdAttrs = @{
        (id)kCTFontAttributeName: (__bridge id)mdFont,
        (id)kCTForegroundColorAttributeName: (__bridge id)[[NSColor whiteColor] CGColor],
    };
    NSAttributedString *mdStr = [[NSAttributedString alloc] initWithString:@"MD" attributes:mdAttrs];
    CTLineRef mdLine = CTLineCreateWithAttributedString((__bridge CFAttributedStringRef)mdStr);
    CGRect mdBounds = CTLineGetBoundsWithOptions(mdLine, 0);
    CGFloat mdX = (size - mdBounds.size.width) / 2 - mdBounds.origin.x;
    CGFloat mdY = size * 0.42;
    CGContextSetTextPosition(ctx, mdX, mdY);
    CTLineDraw(mdLine, ctx);
    CFRelease(mdLine);
    CFRelease(mdFont);

    // ">_" text — monospace, accent color (light blue), lower portion
    CTFontRef promptFont = CTFontCreateWithName(CFSTR("Menlo-Bold"), size * 0.22, NULL);
    CGFloat accentColor[] = {0.4, 0.75, 0.95, 1.0}; // light blue
    CGColorSpaceRef accentSpace = CGColorSpaceCreateDeviceRGB();
    CGColorRef accent = CGColorCreate(accentSpace, accentColor);
    CGColorSpaceRelease(accentSpace);

    NSDictionary *promptAttrs = @{
        (id)kCTFontAttributeName: (__bridge id)promptFont,
        (id)kCTForegroundColorAttributeName: (__bridge id)accent,
    };
    NSAttributedString *promptStr = [[NSAttributedString alloc] initWithString:@">_" attributes:promptAttrs];
    CTLineRef promptLine = CTLineCreateWithAttributedString((__bridge CFAttributedStringRef)promptStr);
    CGRect promptBounds = CTLineGetBoundsWithOptions(promptLine, 0);
    CGFloat promptX = (size - promptBounds.size.width) / 2 - promptBounds.origin.x;
    CGFloat promptY = size * 0.14;
    CGContextSetTextPosition(ctx, promptX, promptY);
    CTLineDraw(promptLine, ctx);
    CFRelease(promptLine);
    CFRelease(promptFont);
    CGColorRelease(accent);

    // Convert to NSImage and set as dock icon
    CGImageRef cgImage = CGBitmapContextCreateImage(ctx);
    CGContextRelease(ctx);
    CGPathRelease(path);

    if (cgImage) {
        NSImage *image = [[NSImage alloc] initWithCGImage:cgImage size:NSMakeSize(size, size)];
        [NSApp setApplicationIconImage:image];
        CGImageRelease(cgImage);
    }
}
*/
import "C"

func setDockIcon() {
	C.guiSetDockIcon()
}

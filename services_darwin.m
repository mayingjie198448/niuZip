#import <Cocoa/Cocoa.h>

extern void goHandleExtractService(char* path);
extern void goHandleCompressService(char* path);

@interface NiuZipServicesProvider : NSObject
- (void)extractHereService:(NSPasteboard *)pboard userData:(NSString *)userData error:(NSString **)error;
- (void)compressToZipService:(NSPasteboard *)pboard userData:(NSString *)userData error:(NSString **)error;
@end

@implementation NiuZipServicesProvider

- (void)extractHereService:(NSPasteboard *)pboard userData:(NSString *)userData error:(NSString **)error {
    // 1. Try reading modern NSURL objects
    NSArray *urls = [pboard readObjectsForClasses:@[[NSURL class]] options:@{}];
    if (urls && [urls count] > 0) {
        for (NSURL *url in urls) {
            if (url && [url isFileURL]) {
                goHandleExtractService((char*)[[url path] UTF8String]);
            }
        }
        return;
    }
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wdeprecated-declarations"
    // 2. Fallback to NSFilenamesPboardType
    id plist = [pboard propertyListForType:NSFilenamesPboardType];
#pragma clang diagnostic pop
    if ([plist isKindOfClass:[NSArray class]]) {
        for (NSString *path in plist) {
            if ([path isKindOfClass:[NSString class]]) {
                goHandleExtractService((char*)[path UTF8String]);
            }
        }
    }
}

- (void)compressToZipService:(NSPasteboard *)pboard userData:(NSString *)userData error:(NSString **)error {
    // 1. Try reading modern NSURL objects
    NSArray *urls = [pboard readObjectsForClasses:@[[NSURL class]] options:@{}];
    if (urls && [urls count] > 0) {
        for (NSURL *url in urls) {
            if (url && [url isFileURL]) {
                goHandleCompressService((char*)[[url path] UTF8String]);
            }
        }
        return;
    }
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wdeprecated-declarations"
    // 2. Fallback to NSFilenamesPboardType
    id plist = [pboard propertyListForType:NSFilenamesPboardType];
#pragma clang diagnostic pop
    if ([plist isKindOfClass:[NSArray class]]) {
        for (NSString *path in plist) {
            if ([path isKindOfClass:[NSString class]]) {
                goHandleCompressService((char*)[path UTF8String]);
            }
        }
    }
}

@end

void initMacServices(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        NiuZipServicesProvider *provider = [[NiuZipServicesProvider alloc] init];
        [NSApp setServicesProvider:provider];
        NSUpdateDynamicServices();
    });
}

#import <Cocoa/Cocoa.h>

extern void goHandleExtractService(char* path);
extern void goHandleCompressService(char* path);

@interface NiuZipServicesProvider : NSObject
- (void)extractHereService:(NSPasteboard *)pboard userData:(NSString *)userData error:(NSString **)error;
- (void)compressToZipService:(NSPasteboard *)pboard userData:(NSString *)userData error:(NSString **)error;
@end

@implementation NiuZipServicesProvider

- (void)extractHereService:(NSPasteboard *)pboard userData:(NSString *)userData error:(NSString **)error {
    NSArray *classes = @[[NSURL class]];
    NSDictionary *options = @{};
    NSArray *urls = [pboard readObjectsForClasses:classes options:options];
    for (NSURL *url in urls) {
        if (url && [url isFileURL]) {
            goHandleExtractService((char*)[[url path] UTF8String]);
        }
    }
}

- (void)compressToZipService:(NSPasteboard *)pboard userData:(NSString *)userData error:(NSString **)error {
    NSArray *classes = @[[NSURL class]];
    NSDictionary *options = @{};
    NSArray *urls = [pboard readObjectsForClasses:classes options:options];
    for (NSURL *url in urls) {
        if (url && [url isFileURL]) {
            goHandleCompressService((char*)[[url path] UTF8String]);
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

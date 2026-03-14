#import <AppKit/AppKit.h>
#include "spellcheck_darwin.h"
#include <stdlib.h>
#include <string.h>

SpellCheckResult spellcheckCheck(const char *text, int textLen) {
    @autoreleasepool {
        SpellCheckResult result = {0};
        if (text == NULL || textLen == 0) return result;

        NSString *str = [[NSString alloc]
            initWithBytes:text
            length:textLen
            encoding:NSUTF8StringEncoding];
        if (str == nil) return result;

        NSSpellChecker *checker = [NSSpellChecker sharedSpellChecker];
        NSInteger strLen = str.length;

        // First pass: count misspelled ranges.
        int capacity = 8;
        int count = 0;
        SpellRange *ranges = (SpellRange *)malloc(
            capacity * sizeof(SpellRange));

        NSInteger offset = 0;
        while (offset < strLen) {
            NSRange checkRange = NSMakeRange(offset,
                strLen - offset);
            NSRange misspelled = [checker
                checkSpellingOfString:str
                startingAt:offset
                language:nil
                wrap:NO
                inSpellDocumentWithTag:0
                wordCount:NULL];
            if (misspelled.location == NSNotFound ||
                misspelled.length == 0) {
                break;
            }

            // Convert UTF-16 range to UTF-8 byte range.
            NSRange utf16Range = misspelled;
            NSString *prefix = [str substringToIndex:
                utf16Range.location];
            NSString *word = [str substringWithRange:
                utf16Range];
            int startByte = (int)[prefix
                lengthOfBytesUsingEncoding:NSUTF8StringEncoding];
            int lenBytes = (int)[word
                lengthOfBytesUsingEncoding:NSUTF8StringEncoding];

            if (count >= capacity) {
                capacity *= 2;
                ranges = (SpellRange *)realloc(ranges,
                    capacity * sizeof(SpellRange));
            }
            ranges[count].startByte = startByte;
            ranges[count].lenBytes = lenBytes;
            count++;

            offset = misspelled.location + misspelled.length;
        }

        result.ranges = ranges;
        result.count = count;
        return result;
    }
}

SuggestResult spellcheckSuggest(const char *text, int textLen,
    int startByte, int lenBytes) {
    @autoreleasepool {
        SuggestResult result = {0};
        if (text == NULL || textLen == 0) return result;

        NSString *str = [[NSString alloc]
            initWithBytes:text
            length:textLen
            encoding:NSUTF8StringEncoding];
        if (str == nil) return result;

        // Convert byte offset to UTF-16 range.
        NSString *prefix = [[NSString alloc]
            initWithBytes:text
            length:startByte
            encoding:NSUTF8StringEncoding];
        NSString *word = [[NSString alloc]
            initWithBytes:text + startByte
            length:lenBytes
            encoding:NSUTF8StringEncoding];
        if (prefix == nil || word == nil) return result;

        NSRange utf16Range = NSMakeRange(prefix.length,
            word.length);

        NSSpellChecker *checker = [NSSpellChecker sharedSpellChecker];
        NSArray<NSString *> *guesses = [checker
            guessesForWordRange:utf16Range
            inString:str
            language:nil
            inSpellDocumentWithTag:0];

        if (guesses == nil || guesses.count == 0) return result;

        int count = (int)guesses.count;
        result.suggestions = (char **)malloc(
            count * sizeof(char *));
        result.count = count;
        for (int i = 0; i < count; i++) {
            const char *s = [guesses[i] UTF8String];
            result.suggestions[i] = strdup(s);
        }
        return result;
    }
}

void spellcheckLearn(const char *word) {
    @autoreleasepool {
        if (word == NULL) return;
        NSString *str = [NSString stringWithUTF8String:word];
        if (str == nil) return;
        [[NSSpellChecker sharedSpellChecker] learnWord:str];
    }
}

void spellcheckFreeResult(SpellCheckResult r) {
    free(r.ranges);
}

void spellcheckFreeSuggestResult(SuggestResult r) {
    for (int i = 0; i < r.count; i++) {
        free(r.suggestions[i]);
    }
    free(r.suggestions);
}

#ifndef _CS_STRING
#define _CS_STRING 1
#include <stdint.h>
#include <assert.h>
#include <wchar.h>
#include <string.h>

#include "shared.h"

typedef struct PACK(csString {
    uint32_t stringSz;
    wchar_t stringData[0x30];
}) csString;

size_t wchar_strlen(const wchar_t* str);
csString make_csstr(wchar_t* str);
//#define make_csstr(str) (csString){ .stringSz = (uint32_t)wchar_strlen(str), .stringData = str }
#define get_size(cstr) ((sizeof(uint32_t) + ((cstr.stringSz) * sizeof(wchar_t)))-1)
#define get_size_ptr(cstr) ((sizeof(uint32_t) + ((cstr->stringSz) * sizeof(wchar_t)))-1)


//#define swap(mem, src_str, dst_str) _swap(mem, &make_csstr(src_str), &make_csstr(dst_str));


#endif

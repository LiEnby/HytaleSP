#include "cs_string.h"

size_t wchar_strlen(const wchar_t* str) {
    size_t len = 0;
    do {
        len++;
    } while (*str++ != L'\0');
    return len - 1;
}

csString make_csstr(wchar_t* str) {
    csString cstr = { 0 };
    cstr.stringSz = (uint32_t)wchar_strlen(str);
    memcpy(cstr.stringData, str, wchar_strlen(str)*sizeof(wchar_t));
    return cstr;
}

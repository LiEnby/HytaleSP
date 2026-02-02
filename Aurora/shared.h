#ifndef _SHARED_H
#define _SHARED_H 1
#include <stdint.h>
#include <wchar.h>
#include <assert.h>

static_assert(sizeof(wchar_t) == 2, "sizeof wchar is not 2 byte; try using -fshort-wchar");

typedef struct modinfo {
	uint8_t* start;
	size_t sz;
} modinfo;


#ifdef __GNUC__
#define PACK( declaration ) declaration __attribute__((__packed__))
#elif _MSC_VER
#define PACK( declaration ) __pragma(pack(push, 1) ) declaration __pragma(pack(pop))
#endif

int get_prot(void* addr);
int change_prot(uintptr_t addr, int newProt);
modinfo get_base();
int get_rw_perms();

// linux:  48 8D ?? ?? E8 ?? ?? ?? 00 80 ?? ?? 00 0F 84
// windows: 48 8D ?? ?? ?? E8 ?? ?? ?? ?? 80 ?? ?? ?? 00 0F 84
#define ISDEBUG_PATTERN_LINUX (mem[0] == 0x48 && mem[1] == 0x8D && mem[4] == 0xE8 && mem[8] == 0x00 && mem[9] == 0x80 && mem[12] == 0x00 && mem[13] == 0x0F && mem[14] == 0x84)
#define ISDEBUG_PATTERN_WINDOWS (mem[0] == 0x48 && mem[1] == 0x8D && mem[5] == 0xE8 && mem[10] == 0x80 && mem[14] == 0x00 && mem[15] == 0x0F && mem[16] == 0x84)

// windows: 00 E8 ?? ?? ?? ?? 80 ?? ?? ?? 00 00 00 00 0F 85 ?? ?? ?? ?? 48 
#define SETAUTH_PATTERN_WINDOWS (mem[0x0] == 0x0 && mem[0x1] == 0xe8 && mem[0x6] == 0x80 && mem[0xa] == 0x0 && mem[0xb] == 0x0 && mem[0xc] == 0x0 && mem[0xd] == 0x0 && mem[0xe] == 0xf && mem[0xf] == 0x85 && mem[0x14] == 0x48)


#ifdef __linux__
#define ISDEBUG_PATTERN_PLATFORM ISDEBUG_PATTERN_LINUX
#elif _WIN32
#define ISDEBUG_PATTERN_PLATFORM ISDEBUG_PATTERN_WINDOWS
#define SETAUTH_PATTERN_PLATFORM SETAUTH_PATTERN_WINDOWS
#endif

#endif

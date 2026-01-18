// dllmain.cpp : Defines the entry point for the DLL application.
#define WIN32_MEAN_AND_LEAN 1
#include <Windows.h>
#include <Psapi.h>
#include <stdint.h>
#include <stdio.h>

void createDebugConsole() {
    FILE* fDummy;
    AllocConsole();
    freopen_s(&fDummy, "CONIN$", "r", stdin);
    freopen_s(&fDummy, "CONOUT$", "w", stderr);
    freopen_s(&fDummy, "CONOUT$", "w", stdout);
}

typedef struct csString {
    uint32_t stringSz;
    wchar_t stringData[0x100];
} csString;

#define make_csstr(str) (csString){.stringSz = wcslen(str), .stringData = str}
#define get_size(cstr) ((sizeof(uint32_t) + (cstr.stringSz * sizeof(wchar_t))) + 1)
#define get_size_ptr(cstr) ((sizeof(uint32_t) + (cstr->stringSz * sizeof(wchar_t))) + 1)

static int swaps = 0;
static int totalSwaps = 6;

#define find(mem, src_str, code) do { csString csSrc = make_csstr(src_str); \
                                        if( memcmp(mem, &csSrc, get_size(csSrc)) == 0 ) { \
                                                code \
                                        } \
                                    } while(0)

#define swap(mem, src_str, dst_str) find((uint8_t*)mem, src_str, csString csNew = make_csstr(dst_str); overwrite((csString*)mem, &csNew); swaps++; )


void overwrite(csString* dst, csString* src) {
    int _ = 0;
    MEMORY_BASIC_INFORMATION mbi = { 0 };
    VirtualQuery(dst, &mbi, sizeof(MEMORY_BASIC_INFORMATION));

    if (!VirtualProtect(mbi.BaseAddress, mbi.RegionSize, PAGE_READWRITE, &mbi.Protect)) return;

    wprintf(L"Overwriting: (%p)%s -> %s\n", dst->stringData, dst->stringData, src->stringData);
    memcpy(dst, src, get_size_ptr(src));

    if (!VirtualProtect(mbi.BaseAddress, mbi.RegionSize, mbi.Protect, &_)) return;
}



void allowOfflineInOnline(uint8_t* src) {
    int _ = 0;
    // 48 8D ?? ?? ?? ?? ?? ?? ?? ?? 80 ?? ?? 40 00 0F
    if (src[0] == 0x48 && src[1] == 0x8D && src[10] == 0x80 && src[13] == 0x40 && src[14] == 0x00 && src[15] == 0x0F) {
        printf("Found debug check: %p\n", src);

        MEMORY_BASIC_INFORMATION mbi = { 0 };
        VirtualQuery(src, &mbi, sizeof(MEMORY_BASIC_INFORMATION));
        if (!VirtualProtect(mbi.BaseAddress, mbi.RegionSize, PAGE_READWRITE, &mbi.Protect)) return;

        printf("Nopping: %p\n", &src[15]);
        memset(&src[15], 0x90, 0x6);

        if (!VirtualProtect(mbi.BaseAddress, mbi.RegionSize, mbi.Protect, &_)) return;
    }

}


void changeServers() {
    MODULEINFO info;
    GetModuleInformation(GetCurrentProcess(), GetModuleHandleA(NULL), &info, sizeof(info));
    
    printf("-------%p\n", info.EntryPoint);
    printf("-------%p\n", info.lpBaseOfDll);
    printf("-------%x\n", info.SizeOfImage);

    uint8_t* memory = (uint8_t*)info.lpBaseOfDll;

    for (uintptr_t i = 0; i < info.SizeOfImage; i++) {
        // allow offline mode while in online mode.
        allowOfflineInOnline(memory + i);

        // swap hytale.com with localhost ..
        swap((memory + i), L"hytale.com", L"127.0.0.1:59313");

        // replace url prefixes to http ..
        swap((memory + i), L"https://account-data.", L"http://");
        swap((memory + i), L"https://sessions.", L"http://");
        swap((memory + i), L"https://telemetry.", L"http://");
        swap((memory + i), L"https://tools.", L"http://");

        // make local servers still run in offline mode
        swap((memory + i), L"authenticated", L"offline");

        if (swaps >= totalSwaps) return;
    }
}

BOOL APIENTRY DllMain( HMODULE hModule,
                       DWORD  ul_reason_for_call,
                       LPVOID lpReserved
                     )
{
    switch (ul_reason_for_call)
    {
    case DLL_PROCESS_ATTACH:
#ifdef _DEBUG
        createDebugConsole();
#endif
        changeServers();
        return TRUE;
    case DLL_THREAD_ATTACH:
    case DLL_THREAD_DETACH:
    case DLL_PROCESS_DETACH:
        break;
    }
    return TRUE;
}


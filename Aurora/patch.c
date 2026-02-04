#include <stdint.h>
#include <stdio.h>

#include "shared.h"
#include "patch.h"
#include "cs_string.h"
#include <stdlib.h>

static int num_swaps = 0;

typedef struct swapEntry {
    csString new;
    csString old;
} swapEntry;

#ifdef _DEBUG
#define print(...) printf(__VA_ARGS__)
#else
#define print(...) /**/
#endif
void overwrite(csString* old, csString* new) {
    int prev = get_prot(old);

    if (change_prot((uintptr_t)old, get_rw_perms()) == 0) {
        int sz = get_size_ptr(new);

        print("overwriting %p with %p\n", old, new);
        memcpy(old, new, sz);
    }

    change_prot((uintptr_t)old, prev);
}


void allowOfflineInOnline(uint8_t* mem) {
    if (PATTERN_PLATFORM) {
        int prev = get_prot(mem);
        // the is online mode and is singleplayer checks
        // are almost right next to eachother, it checks one, then checks the other
        // .. 
        //
        // lea     rcx, [rsp+98h+var_70]
        // call    sub_7FF7E036D780
        // cmp     byte ptr [rsp+98h+var_58], 0
        // jz      loc_7FF7DFEAFAE1
        // mov     rax, [rbx+0C8h]
        // mov     rax, [rax+18h]
        // cmp     qword ptr [rax+0B0h], 0
        // jz      loc_7FF7DFEAF93A
        // .. jz instructions always start with 0F 84 .. ..
        // so we can just scan for that
        // im pretty sure id have to change this approach if i ever wanted to support ARM64 MacOS though ..
        // (or if theres ever a 0F 84 in any of the addresses .. hm but thats a chance of 2^16 :D)

        if (change_prot((uintptr_t)mem, get_rw_perms()) == 0) {
            print("nopping debug check at %p\n", mem);
            for (; (mem[0] != 0x0F && mem[1] != 0x84); mem++); // locate the jz instruction ...
            memset(mem, 0x90, 0x6); // fill with NOP

            print("nopping debug check at %p\n", mem);
            for (; (mem[0] != 0x0F && mem[1] != 0x84); mem++); // locate the next jz instruction ...
            memset(mem, 0x90, 0x6); // fill with NOP
        }


        change_prot((uintptr_t)mem, prev);

    }

}

int needsArgumentModify(const char* program) {
    print("[needsArgumentModify] %s\n", program);
    if (strstr(program, "java") != 0) return 1;
    else return 0;
}

int modifyArgument(const char* program, char* arg) {
    if (strstr(program, "java") != 0) {
        if (strstr(arg, "--session-token=") != 0) {
            return 0; // discard argument
        }
        else if (strstr(arg, "--identity-token=") != 0) {
            return 0; // discard argument
        }
        else if (strstr(arg, "LD_PRELOAD") != 0) {
            return 0; // discard env
        }
        else if (strstr(arg, "--auth-mode=authenticated") != 0) {
            strcpy(arg, "--auth-mode=insecure"); // change to auth-mode=insecure ..
            return 1; // keep argument
        }
        else {
            return 1; // keep argument
        }
    }
    return 1;
}


void swap(uint8_t* mem, csString* old, csString* new) {
    if (memcmp(mem, old, get_size_ptr(old)) == 0) {
        overwrite((csString*)mem, new);
        num_swaps++;
    }
}


void changeServers() {

    swapEntry swaps[] = {
        {.old = make_csstr(L"https://account-data."), .new = make_csstr(L"http://127.0.0")},
        {.old = make_csstr(L"https://sessions."),     .new = make_csstr(L"http://127.0.0")},
        {.old = make_csstr(L"https://telemetry."),    .new = make_csstr(L"http://127.0.0")},
        {.old = make_csstr(L"https://tools."),        .new = make_csstr(L"http://127.0.0")},
        {.old = make_csstr(L"hytale.com"),            .new = make_csstr(L".1:59313")},

    };


    int totalSwaps = (sizeof(swaps) / sizeof(swapEntry));

#ifdef _DEBUG
    // sanity check :
    // 
    // make sure our swaps are always smaller than or the same size as, the original string
    // you can get away with this not being the case on windows as theres extra space, but not on Linux or Mac!
    for (int i = 0; i < totalSwaps; i++) {
        assert(get_size(swaps[i].new) <= get_size(swaps[i].old));
    }
#endif

    modinfo modinf = get_base();
    uint8_t* memory = modinf.start;

    for (size_t i = 0; i < modinf.sz; i++) {
        // allow online mode in offline mode.
        allowOfflineInOnline(&memory[i]);

        for (int sw = 0; sw < totalSwaps; sw++) {
            swap(&memory[i], &swaps[sw].old, &swaps[sw].new);
        }

        if (num_swaps >= totalSwaps) break;
    }


}


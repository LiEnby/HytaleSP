#include <stdio.h>
#include <string.h>
#include <stdint.h>
#include <unistd.h>
#include <stdlib.h>
#include <assert.h>
#include <stddef.h>
#include <sys/mman.h>
#include <errno.h>
#include <wchar.h>

static int swaps = 0;
static int totalSwaps = 6;

int wchar_strlen(const wchar_t * str) {
	size_t len = 0;
	do {
		len++;
	} while(*str++ != L'\0');
	return len-1;
}

typedef struct modinfo {
	uint8_t* start;
	size_t sz;
} modinfo;

int get_change_prot(uintptr_t addr, int newProt) {
	uintptr_t align = (addr - (addr % getpagesize()));
	return mprotect((void*)align, getpagesize(), newProt);
}

int get_prot(void* addr) {
	char line[0x200];

	FILE* fd = fopen("/proc/self/maps", "rb");
	assert(fd != NULL);
	int prot = 0;

	while(fgets(line, sizeof(line), fd) != 0) {
			void* start = 0;
			void* end = 0;
			char sProt[5] = {0};

			sscanf(line, "%p-%p %c%c%c%c ", &start, &end, &sProt[0],&sProt[1],&sProt[2],&sProt[3] );

			if(sProt[3] != 'p') {
				continue;
			};

			if(addr > start && addr < end) {
				if(sProt[0] == 'r') prot |= PROT_READ;
				if(sProt[1] == 'w') prot |= PROT_WRITE;
				if(sProt[2] == 'x') prot |= PROT_EXEC;

				break;
			}

	}

	fclose(fd);
	return prot;
}

modinfo get_base() {

	FILE* fd = fopen("/proc/self/maps", "rb");
	assert(fd != NULL);

	char line[0x200] = {0};
	char bin[0x200] = {0};

	int res = readlink("/proc/self/exe", bin, sizeof(bin));
	assert(res != 0);

	void* begin = 0;
	size_t end = 0;


	while(fgets(line, sizeof(line), fd) != 0) {
		if(strstr(line, bin)) {

			void* sAddr = 0;
			void* sEnd = 0;

			sscanf(line, "%p-%p", &sAddr, &sEnd);
			size_t len = sEnd - sAddr;
			end += len;

			if(begin == 0) begin = sAddr;
			if(sAddr < begin) begin = sAddr;


		}
	};


	fclose(fd);


	return (modinfo){
		.start = begin,
		.sz = end
	};
}


// define structs

typedef struct csString {
    uint32_t stringSz;
    wchar_t stringData[0x100];
} __attribute__((packed, aligned(1))) csString;


#define make_csstr(str) (csString){.stringSz = wchar_strlen(str), .stringData = str}
#define get_size(cstr) ((sizeof(uint32_t) + ((cstr.stringSz) * sizeof(wchar_t)))-1)
#define get_size_ptr(cstr) ((sizeof(uint32_t) + ((cstr->stringSz) * sizeof(wchar_t)))-1)

#define find(mem, src_str, code) do { csString csSrc = make_csstr(src_str); \
                                        if( memcmp(mem, &csSrc, get_size(csSrc)) == 0 ) { \
                                                code \
                                        } \
                                    } while(0)

#define swap(mem, src_str, dst_str) find((uint8_t*)mem, src_str, csString csNew = make_csstr(dst_str); overwrite((csString*)mem, &csNew); swaps++; )

void overwrite(csString* dst, csString* src) {

	int prev = get_prot(dst);

	int b = get_change_prot((uintptr_t)dst, PROT_READ | PROT_WRITE);

	if(b == 0) {
		memcpy(dst, src, get_size_ptr(src));
	} else {
		printf("Failed to change memory protections %s", strerror(errno));
	}

	b = get_change_prot((uintptr_t)dst, prev);

}


void allowOfflineInOnline(uint8_t* src) {

    // linux:  48 8D ?? ?? E8 ?? ?? ?? 00 80 ?? ?? 00 0F 84
    // windows: 48 8D ?? ?? ?? E8 ?? ?? ?? ?? 80 ?? ?? ?? 00 0F 84


    if (src[0] == 0x48 &&
				src[1] == 0x8D &&
				src[4] == 0xE8 &&
				src[8] == 0x00 &&
				src[9] == 0x80 &&
				src[12] == 0x00 &&
				src[13] == 0x0F &&
				src[14] == 0x84) {

				printf("Found debug check @ %p\n", src);

				void* target = &src[13];

				int prev = get_prot(target);
				int b = get_change_prot((uintptr_t)target, PROT_READ | PROT_WRITE);
				if(b == 0){
					memset(target, 0x90, 0x6);
				} else {
					printf("Failed to change memory protections %s\n", strerror(errno));
				}
				b = get_change_prot((uintptr_t)target, prev);

    }


}
__attribute__((constructor)) int run() {
	// cleanup after ourselves.
	unsetenv("LD_PRELOAD");

	printf("Trans rights!\n");

	modinfo inf = get_base();

	uint8_t* memory = inf.start;

	for(size_t i = 0; i < inf.sz; i++) {
		// allow online mode in offline mode.
		allowOfflineInOnline(&memory[i]);

		// replace url prefixes to http ..
		swap(&memory[i], L"https://account-data.", L"http://127.0.0");
		swap(&memory[i], L"https://sessions.", L"http://127.0.0");
		swap(&memory[i], L"https://telemetry.", L"http://127.0.0");
		swap(&memory[i], L"https://tools.", L"http://127.0.0");

		// swap hytale.com with localhost ..
		swap(&memory[i], L"hytale.com", L".1:59313");

    // make local servers still run in offline mode
		swap(&memory[i], L"authenticated", L"offline");

		//if(swaps >= totalSwaps) break;
	}

	printf("TOTAL SWAPS: %d\n", swaps);
}

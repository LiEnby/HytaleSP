#include "cs_string.h"
#include <stdio.h>
#include <string.h>
#include <stdint.h>
#include <stdlib.h>
#include <assert.h>
#include <stddef.h>
#include <errno.h>

#ifdef __linux__
#include <sys/mman.h>
#include <unistd.h>
#include <assert.h>
#elif _WIN32
#define WIN32_MEAN_AND_LEAN 1
#include <windows.h>
#include <psapi.h>
#endif

int change_prot(uintptr_t addr, int newProt) {
#ifdef __linux__ 
	uintptr_t align = (addr - (addr % getpagesize()));
	return mprotect((void*)align, getpagesize(), newProt);
#elif _WIN32
	SYSTEM_INFO sysinf = { 0 };
	MEMORY_BASIC_INFORMATION mbi = { 0 };
	GetSystemInfo(&sysinf);
	VirtualQuery((void*)addr, &mbi, sizeof(MEMORY_BASIC_INFORMATION));
	return VirtualProtect(mbi.BaseAddress, sysinf.dwPageSize, newProt, &mbi.Protect) ? 0 : -1;
#endif
}

int get_prot(void* addr) {
#ifdef __linux__
	char line[0x200];

	FILE* fd = fopen("/proc/self/maps", "rb");
	assert(fd != NULL);
	int prot = 0;

	while (fgets(line, sizeof(line), fd) != 0) {
		void* start = 0;
		void* end = 0;
		char sProt[5] = { 0 };

		sscanf(line, "%p-%p %c%c%c%c ", &start, &end, &sProt[0], &sProt[1], &sProt[2], &sProt[3]);

		if (sProt[3] != 'p') {
			continue;
		};

		if (addr > start && addr < end) {
			if (sProt[0] == 'r') prot |= PROT_READ;
			if (sProt[1] == 'w') prot |= PROT_WRITE;
			if (sProt[2] == 'x') prot |= PROT_EXEC;

			break;
		}

	}

	fclose(fd);
	return prot;
#elif _WIN32
	MEMORY_BASIC_INFORMATION mbi = { 0 };
	VirtualQuery(addr, &mbi, sizeof(MEMORY_BASIC_INFORMATION));
	return mbi.Protect;
#endif
}


modinfo get_base() {
#ifdef __linux__
	FILE* fd = fopen("/proc/self/maps", "rb");
	assert(fd != NULL);

	char line[0x200] = { 0 };
	char bin[0x200] = { 0 };

	int res = readlink("/proc/self/exe", bin, sizeof(bin));
	assert(res != 0);

	void* begin = 0;
	size_t end = 0;


	while (fgets(line, sizeof(line), fd) != 0) {
		if (strstr(line, bin)) {

			void* sAddr = 0;
			void* sEnd = 0;

			sscanf(line, "%p-%p", &sAddr, &sEnd);
			size_t len = sEnd - sAddr;
			end += len;

			if (begin == 0) begin = sAddr;
			if (sAddr < begin) begin = sAddr;


		}
	};
	fclose(fd);
	return (modinfo) {
		.start = begin,
			.sz = end
	};
#elif _WIN32
	MODULEINFO info;
	K32GetModuleInformation(GetCurrentProcess(), GetModuleHandleA(NULL), &info, sizeof(info));

	return (modinfo) {
		.start = info.lpBaseOfDll,
			.sz = info.SizeOfImage
	};
#endif
}

int get_rw_perms() {
#ifdef __linux__
	return PROT_READ | PROT_WRITE;
#elif _WIN32
	return PAGE_READWRITE;
#endif
}

// real entrypoint
void changeServers();

#ifdef __linux__
__attribute__((constructor)) int run() {
	// cleanup after ourselves.
	unsetenv("LD_PRELOAD");
	changeServers();
	return 0;
}
#elif _WIN32
void createConsole() {
	FILE* fDummy;
	AllocConsole();
	freopen_s(&fDummy, "CONIN$", "r", stdin);
	freopen_s(&fDummy, "CONOUT$", "w", stderr);
	freopen_s(&fDummy, "CONOUT$", "w", stdout);
}

BOOL APIENTRY DllMain(HMODULE hModule, DWORD  ul_reason_for_call, LPVOID lpReserved)
{
	switch (ul_reason_for_call)
	{
	case DLL_PROCESS_ATTACH:
#ifdef _DEBUG
		createConsole();
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
#endif
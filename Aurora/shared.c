#include "cs_string.h"
#include "patch.h"
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

#ifdef __linux__ 
int execvpe(const char* filename, char* const argv[], char* const envp[]);
int execve(const char* filename, char* const argv[], char* const envp[]) {

	if (argv == NULL) return -1;
	if (envp == NULL) return -1;

	const char* program = filename;
	if (program == NULL) program = argv[0];
	if (program == NULL) return -1;

	if (needsArgumentModify(program) == 0) {
		return execvpe(filename, argv, envp);
	}

	// count args
	size_t argc = 0;
	for (argc = 0; argv[argc] != NULL; argc++);

	// count envs
	size_t envc = 0;
	for (envc = 0; envp[envc] != NULL; envc++);

	// allocate new argv
	size_t sz = (argc * sizeof(char*)) + 1;
	char** new_argv = malloc(sz);
	memset(new_argv, 0x00, sz);
	size_t new_argc = 0;


	// allocate new envp
	sz = (envc * sizeof(char*)) + 1;
	char** new_envp = malloc(sz);
	memset(new_envp, 0x00, sz);
	size_t new_envc = 0;

	// run modifyArgument on all arguments ..
	for (int i = 0; i < argc; i++) {
		int keep = modifyArgument(program, argv[i]);
		if (keep == 1) {
			new_argv[new_argc] = argv[i];
			new_argc++;
		}
	}

	// run modifyArgument on all environment variabless ..
	for (int i = 0; i < envc; i++) {
		int keep = modifyArgument(program, envp[i]);
		if (keep == 1) {
			new_envp[new_envc] = envp[i];
			new_envc++;
		}
	}

	int ret = execvpe(filename, new_argv, new_envp);

	free(new_argv);
	free(new_envp);

	return ret;
}
#elif _WIN32

static BOOL (WINAPI* CreateProcessW_original)(LPCWSTR lpApplicationName,LPWSTR lpCommandLine,LPSECURITY_ATTRIBUTES lpProcessAttributes,LPSECURITY_ATTRIBUTES lpThreadAttributes,BOOL bInheritHandles,DWORD dwCreationFlags,LPVOID lpEnvironment,LPCWSTR lpCurrentDirectory,LPSTARTUPINFOW lpStartupInfo,LPPROCESS_INFORMATION lpProcessInformation);

BOOL WINAPI CreateProcessW_hook(
	LPCWSTR lpApplicationName,
	LPWSTR lpCommandLine,
	LPSECURITY_ATTRIBUTES lpProcessAttributes,
	LPSECURITY_ATTRIBUTES lpThreadAttributes,
	BOOL bInheritHandles,
	DWORD dwCreationFlags,
	LPVOID lpEnvironment,
	LPCWSTR lpCurrentDirectory,
	LPSTARTUPINFOW lpStartupInfo,
	LPPROCESS_INFORMATION lpProcessInformation
) {
	int wargc = 0;

	wchar_t** wargv = CommandLineToArgvW(lpCommandLine, &wargc);
	if (wargc <= 0) return 0;

	// get program name ...
	wchar_t* longProgram = (wchar_t*)lpApplicationName;
	if (longProgram == NULL) longProgram = wargv[0];

	// get size in utf8
	int len = WideCharToMultiByte(CP_UTF8, 0, longProgram, -1, NULL, 0, NULL, NULL);
	int sz = (len * sizeof(char*)) + 1;

	// convert program name to utf8
	char* program = malloc(sz);
	assert(program != NULL);
	memset(program, 0x00, sz);
	WideCharToMultiByte(CP_UTF8, 0, longProgram, -1, program, len, NULL, NULL);

	// skip everything if its not needed to modify arguments .. 
	if (needsArgumentModify(program) == 0) {
		free(program);

		return CreateProcessW_original(lpApplicationName, lpCommandLine, lpProcessAttributes, lpThreadAttributes, bInheritHandles, dwCreationFlags, lpEnvironment, lpCurrentDirectory, lpStartupInfo, lpProcessInformation);
	}


	// allocate new argv, 
	int new_argc = 0;
	sz = (wargc * sizeof(wchar_t*)) + 1;
	wchar_t** new_wargv = malloc(sz);
	assert(new_wargv != NULL);

	memset(new_wargv, 0x00, sz);

	// check each argument against modifyArgument.

	for (int i = 0; i < wargc; i++) {

		len = WideCharToMultiByte(CP_UTF8, 0, wargv[i], -1, NULL, 0, NULL, NULL);
		sz = (len * sizeof(char)) + 1;
		char* arg = malloc(sz);
		assert(arg != NULL);
		memset(arg, 0x00, sz);

		WideCharToMultiByte(CP_UTF8, 0, wargv[i], -1, arg, len, NULL, NULL);
		int keep = modifyArgument(program, arg);

		if (keep == 1) {
			len = MultiByteToWideChar(CP_UTF8, 0, arg, -1, NULL, 0);
			sz = (len * sizeof(wchar_t)) + 1;
			wchar_t* warg = malloc(sz);
			assert(warg != NULL);
			memset(warg, 0x00, sz);

			MultiByteToWideChar(CP_UTF8, 0, arg, -1, warg, len);

			new_wargv[new_argc] = warg;
			new_argc++;
		}

		free(arg);

	}

	free(program);
	LocalFree(wargv);

	// build lpCommandLine ... 
	lpCommandLine[0] = 0x00;
	for (int i = 0; i < new_argc; i++) {
		if (i > 0) wcscat(lpCommandLine, L" ");

		if (wcschr(new_wargv[i], L' ') != NULL) {
			wcscat(lpCommandLine, L"\"");
			wcscat(lpCommandLine, new_wargv[i]);
			wcscat(lpCommandLine, L"\"");
		}
		else {
			wcscat(lpCommandLine, new_wargv[i]);
		}
	}

	for (int i = 0; new_wargv[i] != NULL; i++) {
		free(new_wargv[i]);
		new_wargv[i] = NULL;
	}
	free(new_wargv);

	// create process ..
	return CreateProcessW_original(lpApplicationName, lpCommandLine, lpProcessAttributes, lpThreadAttributes, bInheritHandles, dwCreationFlags, lpEnvironment, lpCurrentDirectory, lpStartupInfo, lpProcessInformation);
}
#endif



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


#ifdef __linux__
__attribute__((constructor)) int run() {
	// cleanup after ourselves.
	changeServers();
	return 0;
}
#elif _WIN32
void* hook_export_func(const char* targetModuleName, const char* targetFunctionName, void* newPtr) {
	modinfo base = get_base();

	IMAGE_DOS_HEADER* doshdr = (IMAGE_DOS_HEADER*)base.start;
	IMAGE_NT_HEADERS* nthdr = (IMAGE_NT_HEADERS*)(base.start + doshdr->e_lfanew);
	IMAGE_DATA_DIRECTORY dataDirectory = nthdr->OptionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_IMPORT];
	IMAGE_IMPORT_DESCRIPTOR* importDescriptor = (IMAGE_IMPORT_DESCRIPTOR*)(base.start + dataDirectory.VirtualAddress);


	for (int i = 0; importDescriptor[i].Characteristics != 0; i++)
	{
		char* moduleName = (char*)(base.start + importDescriptor[i].Name);

		if (_stricmp(moduleName, targetModuleName) == 0) {
			IMAGE_THUNK_DATA* imageOrigThunkData = (IMAGE_THUNK_DATA*)(base.start + importDescriptor[i].OriginalFirstThunk);
			IMAGE_THUNK_DATA* imageThunkData = (IMAGE_THUNK_DATA*)(base.start + importDescriptor[i].FirstThunk);

			for (int ii = 0; imageOrigThunkData[ii].u1.AddressOfData != 0; ii++) {
				if (imageOrigThunkData[ii].u1.Ordinal & IMAGE_ORDINAL_FLAG) {
					continue;
				}

				IMAGE_IMPORT_BY_NAME* importByName = (IMAGE_IMPORT_BY_NAME*)(base.start + imageOrigThunkData[ii].u1.AddressOfData);

				if (strcmp(importByName->Name, targetFunctionName) == 0) {
					void* origFunction = (void*)imageThunkData[ii].u1.Function;

					// change function ptr
					int prev = get_prot(&imageThunkData[ii]);
					if (change_prot((uintptr_t)&imageThunkData[ii], get_rw_perms()) == 0) {
						(void*)imageThunkData[ii].u1.Function = newPtr;
					}
					change_prot((uintptr_t)&imageThunkData[ii], prev);

					return origFunction;
				}
			}
		}
	}
	return NULL;
}

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
		CreateProcessW_original = hook_export_func("KERNEL32.dll", "CreateProcessW", CreateProcessW_hook);
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

#include <stdio.h>

__declspec(dllexport) int GetUserNameExW(int nfmt, wchar_t* nameBuf, int* sz) {
	if (sz != NULL)
		*sz = 0;

	if (nameBuf != NULL) 
		nameBuf[0] = '\0';

	return 0;
}

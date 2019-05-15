using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.IO;
using System.Runtime.InteropServices;

namespace netRaven
{
    public class TaskHandler
    {
        public enum TaskType
        {
            exit = 0,
            shell = 1,
            processlist = 2,
            download = 3,
            upload = 4,
            inject = 5,
            upgrade = 6
        }

        public static void handleTask(TaskType t, byte[] taskData)
        {
            // TODO: Implement task handling logic
            switch (t)
            {
                case TaskType.exit:
                    break;
                case TaskType.shell:
                    break;
                case TaskType.processlist:
                    break;
                case TaskType.download:
                    break;
                case TaskType.upload:
                    break;
                case TaskType.inject:
                    break;
                case TaskType.upgrade:
                    break;
                default:
                    break;
            }
        }
    }
}

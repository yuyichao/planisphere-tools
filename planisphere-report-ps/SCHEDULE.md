To schedule the planisphere-report.ps1 script using Windows Task Scheduler:

1. Open **Task Scheduler**
2. Select **Create Task** on the **Actions** panel on the right hand side of the window.
3. In the resulting **Create Task** window, under the **General** tab:
    + Give the task a **Name**. (Something like "planisphere-report-ps (hourly)" would do nicely.)
    + Select **Change User or Group**, enter ***"SYSTEM"***, select **Check Names** to confirm, then select **"OK"**.
    + Change **Configure for:** to ***Windows 10*** (or whatever is appropriate for your computer).
4. Under the **Triggers** tab:
    + Select **New**
    + For a daily schedule:
        + Leave **Begin the task** set to ***On a schedule***.
        + Select **Daily** under **Settings**.
        + Change the **Start** if you'd like to start on a prticular day or (more importantly) at a particular hour.
        + Select **OK**.
    + For an hourly schedule:
        + Leave **Begin the task** set to ***On a schedule***.
        + Select **Daily** under **Settings**.
        + Change the **Start** if you'd like to start on a prticular day or (more importantly) at a particular hour.
        + Select the **checkbox** next to **Repeat task every**, leaving the values set to ***1 hour*** for a duration of ***1 day***.
        + Select **OK**.
    + For a DHCP-based trigger:
        + Change **Begin the task**  to ***On an event***.
        + Select **Basic** under **Settings**.
        + Change the **Log** to ***Microsoft-Windows-DHCP Client Events/Admin***.
        + Change the **Source** to ***Dhcp-Client***.
        + Enter an **Event ID** of ***50067***.
        + Select **OK**.
    + For other triggers, consult your favorite search engine or contact OIT Device Engineering.
    > Note: You can have more than one trigger on a Scheduled Task. If using a daily schedule on a laptop, for instance, it is recommended to also add a DHCP-based trigger to reflect changes in network.
5. Under the **Actions** tab:
    + Select **New**
    + Leave **Action** set to ***Start a program***.
    + Enter a **Program/script** of ***powershell.exe***. If you have multiple versions of PowerShell installed on your computer, you may want to **Browse** to the particular version you'd like to use.
    + In **Add arguments (optional)**, enter ***-ExecutionPolicy Bypass -Command ". C:\path\to\your\planisphere-report.ps1; Send-PSReport"***. Make sure to enter *your* correct path and add any options you require before the closing quote (e.g. ...Send-PSReport -Key your-self-report-key").
    + Select **OK**.
6. **IMPORTANT:** Under the **Conditions** tab, 
    + Under the **Power** section, **deselect** the option to ***Start the task only if the computer is on AC power***
    + Under the **Network** section, **select** the option to ***Start only if the following network connection is available***, leaving the option below set to ***Any connection***.
7. Select **OK** at the bottom of the windows to save your Scheduled Task.

> NOTE: Before scheduling your task, it is recommended to open an Administrative Command Prompt (*not* a PowerShell prompt; a *cmd.exe* prompt), and test the command that you intend to run in your Scheduled Task. Error messages that would display at the Command Prompt will not display in your Scheduled Task!

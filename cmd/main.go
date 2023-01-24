package main

import (
//    "encoding/json"
    "fmt"
    "os/exec"
    "log"
    "runtime"
    "github.com/zcalusic/sysinfo"
    "github.com/furlongm/patchman-client-go/clientlib"
)

func main() {
    if runtime.GOOS != "linux" {
        log.Fatal("Error: only linux is supported")
    }

    cmd := exec.Command("ls", "-lah")
    err := cmd.Run()
    if err != nil {
        log.Fatalf("cmd.Run() failed with %s\n", err)
    }

    var si sysinfo.SysInfo
    si.GetSysInfo()

    var hostname =  si.Node.Hostname
    var machine_id = si.Node.MachineID
    var os_name = si.OS.Name
    var os_version = si.OS.Version
    var os_arch = si.Kernel.Architecture
    var kernel = si.Kernel.Release

    fmt.Println("hostname:", string(hostname))
    fmt.Println("machine id:", string(machine_id))
    fmt.Println("OS name: ", string(os_name))
    fmt.Println("OS version:", string(os_version))
    fmt.Println("OS arch:", string(os_arch))
    fmt.Println("Kernel:", string(kernel))

    var rpms, err1 = clientlib.GetInstalledRPMs()
    if err1 != nil {
        log.Fatalf("cmd.Run() failed with %s\n", err1)
    }

    fmt.Println(rpms)
/*
    data, err := json.MarshalIndent(&si, "", "  ")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(string(data))*/
}

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	seccomp "github.com/seccomp/libseccomp-golang"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	// Uncomment to EPERM on os.Open
	id, _ := seccomp.GetSyscallFromName("openat")
	filter, _ := seccomp.NewFilter(seccomp.ActAllow)
	filter.AddRule(id, seccomp.ActErrno.SetReturnCode(int16(syscall.EPERM)))
	filter.Load()

	// Set up the child command to run
	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	// Enable ptrace
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Ptrace: true,
	}
	cmd.Start()
	err := cmd.Wait()
	if err != nil {
		fmt.Printf("Wait returned: %v\n", err)
	}

	pid := cmd.Process.Pid

	for {
		// View the state of the registers
		var regs syscall.PtraceRegs
		err = syscall.PtraceGetRegs(pid, &regs)
		if err != nil {
			break
		}
		// The rax register holds the syscall ID
		syscallID := regs.Orig_rax
		name, _ := seccomp.ScmpSyscall(syscallID).GetName()
		fmt.Printf("[%#x]\t%s\n", syscallID, name)

		// Allow the next syscall to happen...
		err = syscall.PtraceSyscall(pid, 0)
		if err != nil {
			panic(err)
		}

		// ...and wait for the SIGTRAP
		_, err = syscall.Wait4(pid, nil, 0, nil)
		if err != nil {
			panic(err)
		}
	}
}

func usage() {
	fmt.Printf("%s COMMAND [ARGS...]", os.Args[0])
}

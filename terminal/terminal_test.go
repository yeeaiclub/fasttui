package terminal

import (
	"fmt"
	"testing"
	"time"
)

func TestProcessTerminal(t *testing.T) {
	term := NewProcessTerminal()

	fmt.Println("=== ProcessTerminal Test ===")
	fmt.Println("Press Ctrl+C to exit")
	fmt.Println()

	err := term.Start(func(data string) {
		fmt.Printf("Received input: %q\n", data)
	}, func() {
		fmt.Println("Terminal size changed")
	})

	if err != nil {
		fmt.Printf("Start failed: %v\n", err)
		return
	}

	fmt.Println("Terminal started, waiting for input...")
	fmt.Println("Kitty protocol status:", term.IsKittyProtocolActive())

	time.Sleep(500 * time.Millisecond)
	fmt.Println("Kitty protocol status:", term.IsKittyProtocolActive())

	term.Write("Press any key to test...\r\n")

	for range 10 {
		time.Sleep(1 * time.Second)
	}

	term.Stop()
	fmt.Println("\nTest completed")
}

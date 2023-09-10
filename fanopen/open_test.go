package fanopen

import "testing"

func TestRun(t *testing.T) {
	// shouldn't error
	input := "https://google.com/"
	err := Run(input)
	if err != nil {
		t.Errorf("fanopen.Run(\"%s\") threw an error: %s", input, err)
	}

	// should error
	input = "xxxxxxxxxxxxxxx"
	err = Run(input)
	if err == nil {
		t.Errorf("Run(\"%s\") did not throw an error as expected", input)
	}
}

func TestStart(t *testing.T) {
	// shouldn't error
	input := "https://google.com/"
	err := Start(input)
	if err != nil {
		t.Errorf("fanopen.Start(\"%s\") threw an error: %s", input, err)
	}

	// shouldn't error
	input = "xxxxxxxxxxxxxxx"
	err = Start(input)
	if err != nil {
		t.Errorf("fanopen.Start(\"%s\") shouldn't even fail on invalid input: %s", input, err)
	}
}

func TestRunWith(t *testing.T) {
	// shouldn't error
	input := "https://google.com/"
	app := "firefox"
	err := RunWith(input, app)
	if err != nil {
		t.Errorf("fanopen.RunWith(\"%s\", \"%s\") threw an error: %s", input, app, err)
	}

	// should error
	app = "xxxxxxxxxxxxxxx"
	err = RunWith(input, app)
	if err == nil {
		t.Errorf("RunWith(\"%s\", \"%s\") did not throw an error as expected", input, app)
	}
}

func TestStartWith(t *testing.T) {
	// shouldn't error
	input := "https://google.com/"
	app := "firefox"
	err := StartWith(input, app)
	if err != nil {
		t.Errorf("fanopen.StartWith(\"%s\", \"%s\") threw an error: %s", input, app, err)
	}

	// shouldn't error
	input = "[<Invalid URL>]"
	err = StartWith(input, app)
	if err != nil {
		t.Errorf("StartWith(\"%s\", \"%s\") shouldn't even fail on invalid input: %s", input, app, err)
	}

}

func TestRunPath(t *testing.T) {
	input := "D:/"
	err := Start(input)
	if err != nil {
		t.Errorf("fanopen.Start(\"%s\") threw an error: %s", input, err)
	}
}

func TestStartPath(t *testing.T) {
	input := "D:/"
	err := Start(input)
	if err != nil {
		t.Errorf("fanopen.Start(\"%s\") threw an error: %s", input, err)
	}
}

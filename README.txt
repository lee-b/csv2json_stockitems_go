Main docs are in doc/overview.txt.  Please take the time to read it, as I
discovered some problems with the test and have documented my workarounds
and reasons for them, etc.  Note, particularly, that I have aimed for
JSON-compatible output, rather than identical output to the examples given.


To get started, run the followng in BASH.  Note that sourcing profile.sh will
alter your GOPATH.

    source profile.sh
    ./build.sh


Then run:

    bin/stockitems_csv_to_json [--verbose] some_input_file.csv some_output_file.json

For example:

	bin/stockitems_csv_to_json --verbose test_scenarios/given_example/input.csv test_scenarios/given_example/output.json

To check this output against the given example, you can run:

    diff test_scenarios/given_example/expected_output.json test_scenarios/given_example/output.json

Substitute colordiff, meld, vimdiff, etc. for diff, as you prefer.


Note that this code includes a backend library. ./build.sh will generate API
docs for it (in doc/API/index.html) using godoc. You can use godoc directly,
too, of course.

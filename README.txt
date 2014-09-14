Main docs are in doc/overview.txt.  Please take the time to read it, as I
discovered some problems with the test and have documented my workarounds
and reasons for them, etc.


To get started, run the followng in BASH.  Note that sourcing profile.sh will
alter your GOPATH.

    source profile.sh
    ./build.sh


Then run:

    bin/stockitems_csv_to_json some_input_file.csv some_output_file.json


./build.sh will also generate API docs (in doc/API/index.html) using godoc.
You can use godoc directly, if you prefer, of course.


go test command:
	cd ..; go test -run 'TestDecodeGNUSyntax'
	cd ..; go test -run 'TestObjdumpLoong64TestDecodeGNUSyntaxdata' -v -timeout 10h 2>&1 | tee -a log
	cd ..; go test -run 'TestObjdumpLoong64Manual' -v -timeout 10h >&1 | tee -a log
	cd ..; go test -run '.*'

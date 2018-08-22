.PHONY : test clean

test: pycap.so
	go test -coverprofile=cover.out ./pycap/... -v
	LD_LIBRARY_PATH=. python -m unittest discover -s test/

pycap/pycap.c: pycap/pycap.pyx pycap/cpycap.pxd
	cython pycap/pycap.pyx

pycap.so: pycap/pycap.c pycap/*.go
	go build -buildmode=c-shared -o pycap.so ./pycap/...

clean:
	rm -rf pycap.so pycap.h build

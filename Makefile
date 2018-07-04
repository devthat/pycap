.PHONY : clean testgo testpy buildcy test

test: testgo testpy

buildcy: libgocap.so libgocap.h pycap/cpycap.pxd pycap/pycap.pyx
	python setup.py build_ext -i

libgocap.so libgocap.h: ./gocap/*.go
	go build -buildmode=c-shared -o libgocap.so ./gocap/...

testgo:
	@go test -coverprofile=cover.out ./gocap/... -v

testpy: buildcy
	LD_LIBRARY_PATH=. python -m unittest discover -s test/

clean:
	rm -rf build/ cover.out *.so libgocap.h pycap/pycap.c

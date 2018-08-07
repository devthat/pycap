.PHONY : cython go test

test: cython go
	go test -coverprofile=cover.out ./pycap/... -v
	LD_LIBRARY_PATH=. python -m unittest discover -s test/

cython:
	cython pycap/pycap.pyx

go:
	go build -buildmode=c-shared -o pycap.so ./pycap/...

clean:
	rm -rf pycap/pycap.c cover.out pycap.so pycap.h



# test: testgo testpy
#
# buildcy: libgocap.so libgocap.h pycap/cpycap.pxd pycap/pycap.pyx
# 	python setup.py build_ext -i
#
# libgocap.so libgocap.h: ./gocap/*.go
# 	go build -buildmode=c-shared -o libgocap.so ./gocap/...
#
# testgo:
# 	@go test -coverprofile=cover.out ./gocap/... -v
#
# testpy: buildcy
# 	LD_LIBRARY_PATH=. python -m unittest discover -s test/
#
# clean:
# 	rm -rf build/ cover.out *.so libgocap.h pycap/*.c dist/ gocap.egg-info/
#

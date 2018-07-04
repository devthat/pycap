from setuptools import setup
from setuptools.extension import Extension
from Cython.Build import cythonize

setup(
    ext_modules=cythonize([Extension("pycap", ["pycap/pycap.pyx"], libraries=["gocap"])])
)

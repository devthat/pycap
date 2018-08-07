#!/usr/bin/env python

from setuptools import setup
from setuptools.extension import Extension

from Cython.Build import cythonize

# module1 = Extension(
#     'demo',
#     # define_macros=[('MAJOR_VERSION', '1'),
#     #                ('MINOR_VERSION', '0')],
#     include_dirs=['/usr/local/include'],
#     libraries=['tcl83'],
#     library_dirs=['/usr/local/lib'],
#     sources=['demo.c'])

module = Extension("pycap", ["pycap/pycap.pyx"], libraries=["gocap"])
setup(
    name="gocap",
    version="0.0.1",
    description="Parsing of tcp-streams from pcap-files",
    long_description="",
    url='',
    author='Jonathan Koch',
    author_email='devthat@mailbox.org',
    classifiers=(
        'Development Status :: 3 - Alpha',
        'Programming Language :: Python',
        'Programming Language :: Python :: 3.5',
        'Intended Audience :: Developers',
        'Topic :: Software Development :: Libraries :: Python Modules',
    ),
    keywords='',
    ext_modules=cythonize([module])
)

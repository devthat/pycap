#!/usr/bin/env python
from setuptools import setup
from setuptools.extension import Extension


setup(
    name="pycap",
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
    setup_requires=['setuptools-golang'],
    ext_modules=[Extension('pycap', ['pycap/interface.go'])],
    build_golang={'root': 'github.com/user/project'},
)

# Tassen: Next Generation HAL API

This documentation is built using the Sphinix tool and can be built by doing the following:

1. Cloning the repo and installing the prerequisite Python libraries:
   
```
git clone https://github.com/opennetworkinglab/tassen.git
cd tassen/docs/whitepaper
make
```

this should also list the different targets that can be built via the `make` command.

```
$ make
source ./doc_venv/bin/activate ;\
sphinx-build -M help "." "_build"  
Sphinx v1.8.5
Please use `make target' where target is one of
  html        to make standalone HTML files
  dirhtml     to make HTML files named index.html in directories
  singlehtml  to make a single large HTML file
  pickle      to make pickle files
  json        to make JSON files
  htmlhelp    to make HTML files and an HTML help project
  qthelp      to make HTML files and a qthelp project
  devhelp     to make HTML files and a Devhelp project
  epub        to make an epub
  latex       to make LaTeX files, you can set PAPER=a4 or PAPER=letter
  latexpdf    to make LaTeX and PDF files (default pdflatex)
  latexpdfja  to make LaTeX files and run them through platex/dvipdfmx
  text        to make text files
  man         to make manual pages
  texinfo     to make Texinfo files
  info        to make Texinfo files and run them through makeinfo
  gettext     to make PO message catalogs
  changes     to make an overview of all changed/added/deprecated items
  xml         to make Docutils-native XML files
  pseudoxml   to make pseudoxml-XML files for display purposes
  linkcheck   to check all external links for integrity
  doctest     to run all doctests embedded in the documentation (if enabled)
  coverage    to run coverage check of the documentation (if enabled)
```

2. To build the html pages run the following:

```
make html
```

the resulting html files can then be found in the `_build/html/index.html` file

3. To build a pdf document:

You'll first need to install some prerequisite packages, on Ubuntu this would be:

```
apt-get install texlive-latex-base texlive-fonts-recommended texlive-fonts-extra texlive-latex-extra 
```

you can then build the pdf via the following command:

```
make latexpdf
```

the resulting pdf file can be found in the `_build/latext` directory



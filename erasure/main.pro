TEMPLATE = app
CONFIG += console
CONFIG -= app_bundle
CONFIG -= qt
VEC=VEC128
LEMON_DIR=../usr/local/include
QMAKE_CXXFLAGS += -lemon -O3 -msse4 -mavx -D$${VEC} -I$${LEMON_DIR}
QMAKE_CFLAGS += -lemon -O3 -msse4 -D$${VEC}
QMAKE_CC = gcc
QMAKE_CXX = g++

SOURCES += \
    Jerasure-1.2A/galois.c \
    Jerasure-1.2A/jerasure.c \
    Jerasure-1.2A/reed_sol.c \
    Jerasure-1.2A/cauchy.c \ 
    main.cpp \
    Example/example.cpp \
    

HEADERS += \
    Jerasure-1.2A/galois.h \
    Jerasure-1.2A/jerasure.h \
    Jerasure-1.2A/reed_sol.h \
    Jerasure-1.2A/cauchy.h \
    Example/example.h \

DISTFILES += \
    README.md \
    mfile

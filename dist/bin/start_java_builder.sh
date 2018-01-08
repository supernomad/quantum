#!/bin/bash
docker run --rm -v "$PWD":/home/gopher/project -v $HOME/.gradle:/home/gopher/.gradle -v "/home/csaide/workspace/supernomad/quantum:/home/gopher/src/github.com/supernomad/quantum" -v /home/csaide/Android:/home/csaide/Android -w /home/gopher/project --name go4droid -it -u root mpl7/go4droid /bin/bash

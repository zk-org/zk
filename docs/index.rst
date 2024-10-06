.. image:: assets/media/zk-black-modern.png
   :align: center
   :width: 300px

.. image:: assets/media/screencast.svg
   :align: center

.. toctree::
   :hidden:
   :titlesonly:

   GitHub <https://github.com/zk-org/zk>
   Neovim Plugin <https://github.com/zk-org/zk-nvim>

   config/index
   notes/index
   tips/index

`zk` is a plain text note-taking tool that leverages the power of the command line. 

Install as below and then... :doc:`get zettling <tips/getting-started>`!

Installation
============

Homebrew:

.. code-block:: sh
   
   brew install zk

   # Or, if you want to be on the bleeding edge:
   brew install --HEAD zk


Nix:

.. code-block:: sh

   # Run zk from Nix store without installing it:
   nix run nixpkgs#zk

   # Or, to install it permanently:
   nix-env -iA zk

Alpine Linux:

.. code-block:: sh

   # `zk` is currently available in the `testing` repositories:
   apk add zk

Arch Linux:

You can install `the zk package <https://archlinux.org/packages/extra/x86_64/zk/>`_ from the official repos.

.. code-block:: sh

   sudo pacman -S zk

Build from scratch:

Make sure you have a working `Go 1.21+ installation <https://golang.org/>`_, then clone the repository:

.. code-block:: sh

   git clone https://github.com/zk-org/zk.git
   cd zk

On macOS / Linux:

.. code-block:: sh

   make
   ./zk -h




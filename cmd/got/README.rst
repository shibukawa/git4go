got - pure go implementation of git
===================================

Usage
-------

It aims to provide git compatible command. Now it supports only the following sub commands:

* ls-tree
* cat-file

Install
--------

.. code-block:: bash

   $ go get -d github.com/shibukawa/got

Contribution
--------------

1. Fork (https://github.com/shibukawa/got/fork)
2. Create a feature branch
3. Commit your changes
4. Rebase your local changes against the master branch
5. Run test suite with the ``go test ./...`` command and confirm that it passes
6. Run ``gofmt -s``
7. Create a new Pull Request

Author
--------------

* `Yoshiki Shibukawa <https://github.com/shibukawa>`_

Thanks
-------------

To implement git compatible code,  I refers the following codes:

* `git <https://git-scm.com/>`_
* `libgit2 <https://libgit2.github.com/>`_
* `node-git-core <https://github.com/tarruda/node-git-core>`_

License
-------------

It copies some code, comment from original `git <https://git-scm.com/>`_ command. So it is licensed same license GPLv2.

Git related algorithms are implemented in `git4go <https://github.com/shibukawa/git4go>`_, it provides `git2go <https://github.com/libgit2/git2go/>`_ compatible library written in golang.
It is translated from libgit2. So git4go is provided under GPLv2 with linking exception like libgit2.


package:
	python scripts/package.py
	pip install -e .

pypi: package
	python setup.py sdist bdist_wheel upload

publish: package pypi

format: black prettier nixpkgs-fmt

black:
	black .

prettier:
	npx prettier --write .

nixpkgs-fmt:
	nixpkgs-fmt .

setup-hooks:
	cp git-hooks/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit

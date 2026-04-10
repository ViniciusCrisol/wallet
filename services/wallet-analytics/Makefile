PYTHON   := python3.12
VENV     := .venv
DBT      := $(VENV)/bin/dbt
PIP      := $(VENV)/bin/pip

.PHONY: setup deps debug run test docs clean

setup: $(VENV)/bin/activate

$(VENV)/bin/activate: requirements.txt
	$(PYTHON) -m venv $(VENV)
	$(PIP) install --upgrade pip
	$(PIP) install -r requirements.txt
	@touch $(VENV)/bin/activate

run:
	$(DBT) run --profiles-dir .

test:
	$(DBT) test --profiles-dir .

docs:
	$(DBT) docs generate --profiles-dir .
	$(DBT) docs serve --profiles-dir .

deps:
	$(DBT) deps --profiles-dir .

clean:
	rm -rf target/ dbt_packages/ logs/ $(VENV)/

debug:
	$(DBT) debug --profiles-dir .

# Inspection Sheet

## Setup
### Environment
#### Dependencies

1. Install required packages
2. Confirm default configurations
* [ ] Packages installed successfully
* [ ] Defaults match specification

#### Environment variables

1. Validate required environment variables are set
* [ ] Variables align with deployment checklist

### Configuration
#### CLI defaults

1. Inspect generated spreadsheet path
* [ ] Output file lands in build/
* [ ] Delimiter is comma

## Execution
### Workflow
#### CLI run

1. Run casemd with sample.md
* [ ] Exit code is 0
* [ ] Spreadsheet file exists

#### Post-run cleanup

1. Remove temporary files from build/
* [ ] No leftover artifacts

### Validation
#### Error handling

1. Run casemd without --input
* [ ] CLI prints actionable error
* [ ] Exit code is 1

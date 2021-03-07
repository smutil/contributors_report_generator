# Contributors Report Generator ![example workflow](https://github.com/smutil/contributors_report_generator/actions/workflows/build-actions.yml/badge.svg)

CLI to generate contributors detail and commit count for given list of GIT repository.

Usage
-----
 step 1. download contributors_report_generator from <a href=https://github.com/smutil/contributors_report_generator/releases>releases</a> and extract the tarball.
 
 step 2. create [config.yml](https://github.com/smutil/rcontributors_report_generator/config.yml). If config.yaml and contributors_report_generator is not in same location, you can provide the config.yml path using --config
 
 step 3. execute the contributors_report_generator as shown below. 
 
 ```
 ./contributors_report_generator --config /path-to-config.yml
 ```
 step 4. contributors_report.xlsx will be generated in same location.

 ![Alt text](docs/images/example_xlsx.png?raw=true "Title")


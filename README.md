# Mychron Data Uploader
Uploads Mychron csv data to Mongodb for further analytics.

### Requirements
- MongoDB database
- Aim Csv Format Mychron data

### Usage

Export your data from Race Studio in Aim Csv Format.

If you want to do a dry run, in your `.env` set `LOAD_SESSION=false` (or anything except `true`).

Run the uploader with a list of the file paths for the csvs you want to upload.

`./uploader myFile.csv myFile2.csv`

### Optional - Visualize Data

I used Grafana with the MongoDB plugin to create a [public dashboard](https://owensmallwood2.grafana.net/public-dashboards/2cf5c574169049b29bcf659992ba7ce9) with some of the lap data.

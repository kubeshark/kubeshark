#!/bin/bash

# exit when any command fails
set -e

dst_folder=$1
echo "dst folder: $dst_folder";

cd $dst_folder/../ui-common
npm i
npm pack
mv up9-mizu-common-0.0.0.tgz $dst_folder

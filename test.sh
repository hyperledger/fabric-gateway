#!/bin/bash

base_dir=$PWD
cd $base_dir/client/go/sdk/scenario
godog ../../../../scenario/features/

cd $base_dir/client/node/sdk
./node_modules/.bin/cucumber-js --require './steps/**/*.js' ../../../scenario/features/*.feature

cd $base_dir

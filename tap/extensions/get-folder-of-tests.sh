kubeshark_test_files_repo=https://github.com/kubeshark/kubeshark-tests-data
kubeshark_test_files_tmp_folder="kubeshark-test-files-tmp"

requested_folder=$1
destination_folder_name=$2
echo "Going to download folder (${requested_folder}) from repo (${kubeshark_test_files_repo}) and save it to local folder (${destination_folder_name})"

echo "Cloning repo to tmp folder (${kubeshark_test_files_tmp_folder})"
git clone ${kubeshark_test_files_repo} ${kubeshark_test_files_tmp_folder} --no-checkout --depth 1 --filter=blob:none --sparse
cd ${kubeshark_test_files_tmp_folder}

echo "Adding sparse checkout folder"
git sparse-checkout add ${requested_folder}


echo "Checkout"
git checkout 

echo "Moving folder to the destination location"
mv ${requested_folder} ../${destination_folder_name}

cd ..
echo "Removing the tmp folder"
rm -rf ${kubeshark_test_files_tmp_folder}

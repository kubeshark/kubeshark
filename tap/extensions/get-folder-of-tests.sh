mizu_test_files_repo=https://github.com/up9inc/mizu-tests-data
mizu_test_files_tmp_folder="mizu-test-files-tmp"

requested_folder=$1
destination_folder_name=$2
echo "Going to download folder (${requested_folder}) from repo (${mizu_test_files_repo}) and save it to local folder (${destination_folder_name})"

echo "Cloning repo to tmp folder (${mizu_test_files_tmp_folder})"
git clone ${mizu_test_files_repo} ${mizu_test_files_tmp_folder} --no-checkout --depth 1 --filter=blob:none --sparse
cd ${mizu_test_files_tmp_folder}

echo "Adding sparse checkout folder"
git sparse-checkout add ${requested_folder}


echo "Checkout"
git checkout 

echo "Moving folder to the destination location"
mv ${requested_folder} ../${destination_folder_name}

cd ..
echo "Removing the tmp folder"
rm -rf ${mizu_test_files_tmp_folder}

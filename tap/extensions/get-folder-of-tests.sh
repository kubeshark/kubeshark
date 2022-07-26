mizu_test_files_repo=https://github.com/up9inc/mizu-tests-data
mizu_test_files_tmp_folder="mizu-test-files-tmp"

requested_folder=$1
destination_folder_name=$2
echo "Going to download folder (${requested_folder}) from repo (${mizu_test_files_repo}) and save it to local fodler (${destination_folder_name}"

git clone ${mizu_test_files_repo} ${mizu_test_files_tmp_folder} --no-checkout --depth 1 --filter=blob:none --sparse
echo "Cloneing repo to tmp folder (${mizu_test_files_tmp_folder})"
cd ${mizu_test_files_tmp_folder}

git sparse-checkout add ${requested_folder}
echo "Adding sparse checkout folder"

git checkout 
echo "Checkout"

mv ${requested_folder} ../${destination_folder_name}
echo "Moving folder to the destination location"

cd ..
echo "Removing the tmp folder"
rm -rf ${mizu_test_files_tmp_folder}

#!/usr/bin/env bash

osName="linux"
case "`uname`" in
 Linux)
  osName="linux"
  ;;
 Darwin)
 osName="mac"
  echo "mac system"
  ;;
 *)
  echo `uname`
  # exit 1
 ;;
esac

for element in `ls .`
do
	dir_or_file=$element
	if [ -d $dir_or_file ]
	then
		echo -e "protoc -I . $dir_or_file/$dir_or_file.proto --go_out=plugins=grpc:.\n"
		`protoc -I . $dir_or_file/$dir_or_file.proto --go_out=plugins=grpc:.`
		if [ ${osName} == "linux" ]; then
			`sed -i 's/AASMessage "pb\/AASMessage"/AASMessage "github.com\/lyesteven\/go-framework\/pb\/AASMessage"/' pb/$dir_or_file/$dir_or_file.pb.go`
			`sed -i 's/DALMessage "pb\/DALMessage"/DALMessage "github.com\/lyesteven\/go-framework\/pb\/DALMessage"/' pb/$dir_or_file/$dir_or_file.pb.go`
			`sed -i 's/ComMessage "pb\/ComMessage"/ComMessage "github.com\/lyesteven\/go-framework\/pb\/ComMessage"/' pb/$dir_or_file/$dir_or_file.pb.go`

		else
			`sed -i '' 's/AASMessage "pb\/AASMessage"/AASMessage "github.com\/lyesteven\/go-framework\/pb\/AASMessage"/' pb/$dir_or_file/$dir_or_file.pb.go`
			`sed -i '' 's/DALMessage "pb\/DALMessage"/DALMessage "github.com\/lyesteven\/go-framework\/pb\/DALMessage"/' pb/$dir_or_file/$dir_or_file.pb.go`
			`sed -i '' 's/ComMessage "pb\/ComMessage"/ComMessage "github.com\/lyesteven\/go-framework\/pb\/ComMessage"/' pb/$dir_or_file/$dir_or_file.pb.go`
		fi

		if [ ! -d "../pb/$dir_or_file" ]; then
			`mkdir -p ../pb/$dir_or_file`
		fi
		`mv pb/$dir_or_file/$dir_or_file.pb.go ../pb/$dir_or_file/`
	fi
done

if [ -d "pb" ]; then
	`rm -fr pb`
fi

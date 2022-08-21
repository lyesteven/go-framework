#!/bin/sh
TOPPATH=$GOPATH/src/gworld/git/GoFrameWork
ONEPKG=gworld/git/third-package/satori/go.uuid
OLDPWD=$PWD
INCFILE=/tmp/gworldBaseModFiles
TEMPFILE=__tmpdata

function cleanMod()
{
	rm -f "$INCFILE"
	`find $GOPATH/src/gworld/git -type f \( -name "go.mod" -or -name "go.sum" \) -exec rm -f {} \;`
}

function gomodinit()
{
	#1st param, the dir name
	currentpath=`pwd`
	cd $1
	childpath=`pwd`
	if [[ ! -e "go.mod" ]]; then
		go mod init
		echo ${childpath##$GOPATH'/src/'} >> $INCFILE
	fi

	cd $currentpath
}
function checkBaseMod()
{
        #1st param, the dir name
        #2nd param, the aligning space
		#3rd param, the parent dir name

		#echo "   =========begin checkBaseMod($1)"
		# 第一次执行 (没有$INCFILE)
		if [[ ! -e "$INCFILE" ]];then
			`find $TOPPATH -name "go.mod" | xargs rm -f`
			`find $TOPPATH -name "go.sum" | xargs rm -f`
			#echo $ONEPKG >> $INCFILE
			gomodinit "$GOPATH/src/$ONEPKG"
		fi

        for file in `ls $1`;
        do
				parent=$3
                if [ -d "$1/$file" ]; then					
                    #echo "$2$file --$parent"
					if [[ "$parent" = "services" || "$parent" = "pb" ]] ; then
                    	#echo "$2$file --$parent"
						gomodinit "$1/$file"
					elif [[ "$parent" = "common" && "$file" != "utils" ]]; then
                    	#echo "$2$file --$parent"
						gomodinit "$1/$file"
					elif [ "$file" = "grpcUtility" ] ; then
                    	#echo "$2$file --$parent"
						gomodinit "$1/$file"
					else
                    	checkBaseMod "$1/$file" "   $2" $file
					fi
                fi
        done
		#echo "   =============end checkBaseMod($1)"
}

function tidyModFile()
{
	#echo -e "\n\n`cat go.mod`"
	for name in `cat go.mod |grep "gworld" | awk '{print $1}' | sort |uniq -c | sort -n | awk '$1<2 {print $2}'|awk -F'/' '{print$(NF-1)"\\\/"$NF}'`
	do
		#echo "============== $name"
		`sed -i "/^module /b; /$name/d" go.mod`
	done
	#echo -e "\n\n`cat go.mod`"
}

function depend()
{
	#echo "begin depend()"
	#删除 "replace (" 后的所有行
	`sed -i '/replace (/,$d' go.mod`
	#删除所有 "gworld/git/...." 依赖
	`sed -i "/^module /b; /gworld\/git/d" go.mod`

	currentpath=`pwd`
	for name in `go mod tidy 2>&1 | grep gworld | grep -E '(package|malformed)' | awk -F':' '{print $1}'` 
	do 
		echo -e "\t $name => ${GOPATH}/src/$name " >> $TEMPFILE
		if [[ ! -e "${GOPATH}/src/$name/go.mod" ]]; then
			cd ${GOPATH}/src/$name
			go mod init
			cd $currentpath
		fi
	done

	for name in `cat $INCFILE` 
	do 
		echo -e "\t $name => ${GOPATH}/src/$name " >> $TEMPFILE 
	done

	echo -e "\nreplace (" >> go.mod
	cat $TEMPFILE >> go.mod
	echo ")" >> go.mod

	rm -f $TEMPFILE
	#echo "end depend()"
}

if [[ "$1" = "init" ]];then
	checkBaseMod $TOPPATH ""
elif [[ "$1" = "depends" && "${GO111MODULE}" != "off" ]]; then
	depend
	`go mod tidy`
	tidyModFile
elif [[ "$1" = "cleanmod" ]];then
	cleanMod
fi

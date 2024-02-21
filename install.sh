#!/bin/sh

EXE_NAME=kubeshark
ALIAS_NAME=ks
PROG_NAME=Kubeshark
INSTALL_PATH=/usr/local/bin/$EXE_NAME
ALIAS_PATH=/usr/local/bin/$ALIAS_NAME
REPO=https://github.com/kubeshark/kubeshark
OS=$(echo $(uname -s) | tr '[:upper:]' '[:lower:]')
ARCH=$(echo $(uname -m) | tr '[:upper:]' '[:lower:]')
SUPPORTED_PAIRS="linux_amd64 linux_arm64 darwin_amd64 darwin_arm64"

ESC="\033["
F_DEFAULT=39
F_RED=31
F_GREEN=32
F_YELLOW=33
B_DEFAULT=49
B_RED=41
B_BLUE=44
B_LIGHT_BLUE=104

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
fi

if [ "$ARCH" = "aarch64" ]; then
    ARCH="arm64"
fi

echo $SUPPORTED_PAIRS | grep -w -q "${OS}_${ARCH}"

if [ $? != 0 ] ; then
	echo "\n${ESC}${F_RED}m🛑 Unsupported OS \"$OS\" or architecture \"$ARCH\". Failed to install $PROG_NAME.${ESC}${F_DEFAULT}m"
    echo "${ESC}${B_RED}mPlease report 🐛 to $REPO/issues${ESC}${F_DEFAULT}m"
	exit 1
fi

echo "\n🦈 ${ESC}${F_DEFAULT};${B_BLUE}m Started to download $PROG_NAME ${ESC}${B_DEFAULT};${F_DEFAULT}m"

if curl -# --fail -Lo $EXE_NAME ${REPO}/releases/latest/download/${EXE_NAME}_${OS}_${ARCH} ; then
    chmod +x $PWD/$EXE_NAME
    echo "\n${ESC}${F_GREEN}m⬇️  $PROG_NAME is downloaded into $PWD/$EXE_NAME${ESC}${F_DEFAULT}m"
else
    echo "\n${ESC}${F_RED}m🛑 Couldn't download ${REPO}/releases/latest/download/${EXE_NAME}_${OS}_${ARCH}\n\
  ⚠️  Check your internet connection.\n\
  ⚠️  Make sure 'curl' command is available.\n\
  ⚠️  Make sure there is no directory named '${EXE_NAME}' in ${PWD}\n\
${ESC}${F_DEFAULT}m"
    echo "${ESC}${B_RED}mPlease report 🐛 to $REPO/issues${ESC}${F_DEFAULT}m"
    exit 1
fi

use_cmd=$EXE_NAME
printf "Do you want to install system-wide? Requires sudo 😇 (y/N)? "
old_stty_cfg=$(stty -g)
stty raw -echo ; answer=$(head -c 1) ; stty $old_stty_cfg
if echo "$answer" | grep -iq "^y" ;then
    echo "$answer"
    sudo mv ./$EXE_NAME $INSTALL_PATH || exit 1
    echo "${ESC}${F_GREEN}m$PROG_NAME is installed into $INSTALL_PATH${ESC}${F_DEFAULT}m\n"

	ls $ALIAS_PATH >> /dev/null 2>&1
	if [ $? != 0 ] ; then
		printf "Do you want to add 'ks' alias for Kubeshark? (y/N)? "
		old_stty_cfg=$(stty -g)
		stty raw -echo ; answer=$(head -c 1) ; stty $old_stty_cfg
		if echo "$answer" | grep -iq "^y" ; then
			echo "$answer"
			sudo ln -s $INSTALL_PATH $ALIAS_PATH

			use_cmd=$ALIAS_NAME
		else
			echo "$answer"
		fi
	else
		use_cmd=$ALIAS_NAME
	fi
else
	echo "$answer"
	use_cmd="./$EXE_NAME"
fi

echo "${ESC}${F_GREEN}m✅ You can use the ${ESC}${F_DEFAULT};${B_LIGHT_BLUE}m $use_cmd ${ESC}${B_DEFAULT};${F_GREEN}m command now.${ESC}${F_DEFAULT}m"
echo "\n${ESC}${F_YELLOW}mPlease give us a star 🌟 on ${ESC}${F_DEFAULT}m$REPO${ESC}${F_YELLOW}m if you ❤️  $PROG_NAME!${ESC}${F_DEFAULT}m"

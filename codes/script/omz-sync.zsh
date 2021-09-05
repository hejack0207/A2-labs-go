#!/usr/bin/zsh
hash | grep --color=never 'go/scripts/.*\.go$' | awk -F= '{print $2}' | sed -re 's#(/home/he/.oh-my-zsh/plugins)/(.*)/(.*\.go)#cp \1/\2/\3 \3#' | bash -s

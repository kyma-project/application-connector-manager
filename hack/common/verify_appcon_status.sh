#!/bin/bash

function get_appcon_status () {
	local number=1
	echo "test test"
	while [[ $number -le 100 ]] ; do
		echo ">--> checking application-connector status #$number"
		local STATUS=$(kubectl get applicationconnector -n kyma-system applicationconnector-sample -o jsonpath='{.status.state}')
		echo "application-connector status: ${STATUS:='UNKNOWN'}"
		[[ "$STATUS" == "Ready" ]] && return 0
		sleep 5
        	((number = number + 1))
	done

	kubectl get all --all-namespaces
	exit 1
}

get_appcon_status

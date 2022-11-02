// +build darwin
package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

int
SetActivationPolicy(void) {
    [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
    return 0;
}
*/
import "C"
import (
	log "github.com/sirupsen/logrus"
)

func SetActivationPolicy() {
	log.Infoln("Setting activation Policy")
	C.SetActivationPolicy()
}

//go:build tools

package main

import (
	_ "github.com/golang/mock/gomock"
	_ "github.com/golang/mock/mockgen/model"
	_ "github.com/stretchr/testify"
	_ "github.com/stretchr/testify/assert"
)

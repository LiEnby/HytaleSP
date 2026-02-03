package main

import (
	"context"

	"github.com/itchio/lake/pools/fspool"
	"github.com/itchio/savior/filesource"
	"github.com/itchio/wharf/pwr"
	"github.com/itchio/wharf/pwr/bowl"
	"github.com/itchio/wharf/pwr/patcher"
	"github.com/itchio/headway/state"
	"fmt"

	_ "github.com/itchio/wharf/decompressors/cbrotli"
)


func applyPatch(source string, destination string, patchFilename string, signatureFilename string) error {

	patchReader, err := filesource.Open(patchFilename);
	if err != nil {
		return err;
	}
	defer patchReader.Close();

	consumer := &state.Consumer{
		OnMessage: func(level string, message string) {
			fmt.Printf("[%s] %s\n", level, message)
		},
	}

	p, err := patcher.New(patchReader, consumer);
	if err != nil {
		return err;
	}

	targetPool := fspool.New(p.GetTargetContainer(), source);

	b, err := bowl.NewFreshBowl(bowl.FreshBowlParams{
		SourceContainer: p.GetSourceContainer(),
				    TargetContainer: p.GetTargetContainer(),
				    TargetPool: targetPool,
				    OutputFolder: destination,
	});

	if err != nil {
		return err;
	}

	err = p.Resume(nil, targetPool, b);
	if err != nil {
		return err;
	}

	sigSource, err := filesource.Open(signatureFilename);
	if err != nil {
		return err;
	}

	signatureInfo, err := pwr.ReadSignature(context.Background(), sigSource);
	err = pwr.AssertValid(destination, signatureInfo);
	if err != nil {
		return err;
	}

	err = pwr.AssertNoGhosts(destination, signatureInfo);
	if err != nil {
		return err;
	}

	return nil;
}

// Copyright 2022 Vitalii Noha vitalii.noga@gmail.com. All rights reserved.

package sgo

func (b *builder) Build(app string) error {
	if err := b.builder.Build(app); err != nil {
		return err
	}
	return nil
}

func (b *builder) Clean(app string, sources *map[string][][]string) error {
	// remove the built files
	if err := b.builder.Clean(app); err != nil {
		return err
	}
	// remove the generated files
	b.coder.Init(*sources)
	if err := b.coder.Clean(app); err != nil {
		return err
	}
	return nil
}

func (b *builder) Generate(app string, sources *map[string][][]string) error {
	b.coder.Init(*sources)
	if err := b.coder.Generate(app); err != nil {
		return err
	}
	return nil
}

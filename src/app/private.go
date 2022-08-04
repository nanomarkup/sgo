// Copyright 2022 Vitalii Noha vitalii.noga@gmail.com. All rights reserved.

package app

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mitchellh/go-ps"
	"github.com/spf13/viper"
)

const (
	// appFileName constant returns package's app file name
	appFileName string = "app.go"
	// depsFileName constant returns name of file with all dependencies
	depsFileName string = "deps.go"
	// appsItemName constant returns a main item name
	appsItemName string = "apps"
	// entryAttrName constant returns an entry attribute name of the application
	entryAttrName string = "entry"
)

type adapter struct {
	code    map[string][]string
	imports *imports
}

type resolver struct {
	application string
	entryPoint  string
	items       map[string]map[string]string
}

type item struct {
	kind     uint
	name     string
	pkg      string
	path     string
	original string
	deps     items
}

type items map[string]item

type alias string

type imports map[string]alias

var itemKind = struct {
	Func   uint
	Struct uint
	String uint
	Number uint
}{
	1,
	2,
	3,
	4,
}

type typeInfo struct {
	Id      string
	Kind    reflect.Kind
	Name    string
	String  string
	PkgPath string
	Fields  []field
	Methods []method
}

type field struct {
	Id        string
	Kind      reflect.Kind
	TypeName  string
	FieldName string
	PkgPath   string
}

type method struct {
	Name string
	In   []field
	Out  []field
}

func getType(types []typeInfo, id string) *typeInfo {
	for _, v := range types {
		if v.Id == id {
			return &v
		}
	}
	return nil
}

func getTypeInfo(list []typeInfo) ([]typeInfo, error) {
	// process all items
	main := []string{}
	imports := map[string]string{}
	if len(list) > 0 {
		impId := 0
		impRef := ""
		itemId := 0
		found := false
		for _, x := range list {
			// the struct and interface types are supported only
			if x.Kind != reflect.Struct && x.Kind != reflect.Interface {
				continue
			}
			// update imports
			itemId++
			impRef, found = imports[x.PkgPath]
			if !found {
				impId++
				impRef = fmt.Sprintf("i%d", impId)
				imports[x.PkgPath] = impRef
			}
			main = append(main, genSerializeType(itemId, impRef, x))
		}
	}
	// populate the import section
	unit := []string{}
	unit = append(unit, "package main\n")
	unit = append(unit, `import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"reflect"
	`)
	if len(imports) > 0 {
		for k, v := range imports {
			unit = append(unit, fmt.Sprintf("\t%s \"%s\"", v, k))
		}
	}
	unit = append(unit, ")\n")
	// populate the main function
	unit = append(unit, `func main() {
	data := []Type{}`)

	if len(main) == 0 {
		unit = append(unit, "\treturn")
	} else {
		unit = append(unit, main...)
	}
	unit = append(unit, `	serialize(data)
}
`)
	// create a temporary folder
	dir, err := ioutil.TempDir("", "sb*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	// add other functions
	typesPath := filepath.Join(dir, "types")
	if len(main) > 0 {
		unit = append(unit, genSerializeFunc(typesPath))
		defer func() {
			os.Remove(typesPath)
		}()
	}
	// generate a main unit and run it
	fp := filepath.Join(dir, "main.go")
	file, err := os.Create(fp)
	if err != nil {
		return nil, err
	}
	defer func() {
		file.Close()
		os.Remove(fp)
	}()

	writer := bufio.NewWriter(file)
	writer.WriteString(strings.Join(unit, "\n"))
	writer.Flush()
	// serialize items
	if _, err = goRun(fp); err != nil {
		return nil, err
	}
	// deserialize items
	if _, err := os.Stat(typesPath); err == nil {
		types, err := ioutil.ReadFile(typesPath)
		if err != nil {
			return nil, err
		}
		var info []typeInfo
		dec := gob.NewDecoder(bytes.NewReader(types))
		if err := dec.Decode(&info); err != nil {
			return nil, err
		} else {
			return info, nil
		}
	}
	return nil, errors.New("cannot collect type details")
}

func isDirEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, nil
}

func isDebugging() bool {
	pid := os.Getppid()
	// We loop in case there were intermediary processes like the gopls language server.
	for pid != 0 {
		p, err := ps.FindProcess(pid)
		if err != nil || p == nil {
			return false
		}
		if p.Executable() == "dlv.exe" {
			return true
		}
		pid = p.PPid()
	}
	return false
}

func goRun(src string) ([]byte, error) {
	args := []string{"run", src}
	cmd := exec.Command("go", args...)
	if isDebugging() {
		// resolve the debugging sb application
		cmd.Dir, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	} else {
		// resolve the testing sb application
		wd := viper.GetString("GOWD")
		if wd != "" {
			cmd.Dir = wd
		}
	}
	out, err := cmd.Output()
	if e, ok := err.(*exec.ExitError); ok {
		return out, fmt.Errorf("%s", e.Stderr)
	} else {
		return out, err
	}
}

func goBuild(src, dst string) error {
	args := []string{"build"}
	if dst != "" {
		args = append(args, "-o", dst)
	}
	args = append(args, src)
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func goClean(src string) error {
	cmd := exec.Command("go", "clean", src)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func readItem(name string, items map[string]map[string]string) (map[string]string, error) {
	if apps, found := items[name]; found {
		return apps, nil
	}
	return nil, fmt.Errorf("The %s item is not found", name)
}

func appendImport(list imports, path string) alias {
	if path == "" || path[0:1] == "." {
		return ""
	}
	item := list[path]
	if item != "" {
		return item
	} else {
		item = alias(fmt.Sprintf("p%d", len(list)+1))
		list[path] = item
		return item
	}
}

func checkApplication(application string) error {
	if application == "" {
		return fmt.Errorf("The application is not specified")
	}
	return nil
}

func genSerializeType(id int, imp string, x typeInfo) string {
	return fmt.Sprintf("\tvar v%d %s.%s\n", id, imp, x.Name) +
		fmt.Sprintf("\tdata = append(data, getType(&v%d))", id)
}

func genSerializeFunc(filePath string) string {
	return `type Field struct {
	Id        string
	Kind      reflect.Kind
	TypeName  string
	FieldName string
	PkgPath   string
}

type Method struct {
	Name string
	In   []Field
	Out  []Field
}

type Type struct {
	Id      string
	Kind    reflect.Kind
	Name    string
	String  string
	PkgPath string
	Fields  []Field
	Methods []Method
}

func getType(v interface{}) Type {
	e := reflect.TypeOf(v).Elem()
	info := Type{
		Id:      fmt.Sprintf("%s.%s", e.PkgPath(), e.Name()),
		Kind:    e.Kind(),
		Name:    e.Name(),
		String:  e.String(),
		PkgPath: e.PkgPath(),
	}
	if e.Kind() == reflect.Struct {
		info.Fields = getFields(e)
		info.Methods = getMethods(reflect.TypeOf(v))
	} else if e.Kind() == reflect.Interface {
		info.Methods = getMethods(e)
	}
	return info
}

func getFields(t reflect.Type) []Field {
	res := []Field{}
	var f reflect.StructField
	for i := 0; i < t.NumField(); i++ {
		f = t.Field(i)
		res = append(res, Field{
			Id:        fmt.Sprintf("%s.%s", f.Type.PkgPath(), f.Type.Name()),
			Kind:      f.Type.Kind(),
			TypeName:  f.Type.Name(),
			FieldName: f.Name,
			PkgPath:   f.Type.PkgPath(),
		})
	}
	return res
}

func getMethods(t reflect.Type) []Method {
	res := []Method{}
	var x Method
	var m reflect.Method
	for i := 0; i < t.NumMethod(); i++ {
		m = t.Method(i)
		x = Method{Name: m.Name}
		// input params
		for n := 0; n < m.Type.NumIn(); n++ {
			ti := m.Type.In(n)
			x.In = append(x.In, Field{
				Id:        fmt.Sprintf("%s.%s", ti.PkgPath(), ti.Name()),
				Kind:      ti.Kind(),
				TypeName:  ti.Name(),
				FieldName: ti.Name(),
				PkgPath:   ti.PkgPath(),
			})
		}
		// output params
		for n := 0; n < m.Type.NumOut(); n++ {
			to := m.Type.Out(n)
			x.Out = append(x.Out, Field{
				Id:        fmt.Sprintf("%s.%s", to.PkgPath(), to.Name()),
				Kind:      to.Kind(),
				TypeName:  to.Name(),
				FieldName: to.Name(),
				PkgPath:   to.PkgPath(),
			})
		}
		res = append(res, x)
	}
	return res
}

func serialize(info []Type) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(info)
	if err == nil {
		file, err := os.Create("` + strings.ReplaceAll(filePath, "\\", "\\\\") + `")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer func() {
			file.Close()
		}()
		writer := bufio.NewWriter(file)
		writer.Write(buf.Bytes())
		writer.Flush()
	} else {
		fmt.Println(err)
	}
}`
}

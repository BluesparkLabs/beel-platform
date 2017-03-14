package main

import (
  "fmt"
  "os"
  "github.com/pelletier/go-toml"
  "strconv"
  "reflect"
  "net/http"
  "io"
)

func msg_error(message string) {
  fmt.Println("[ERROR]", message)
  os.Exit(1)
}

func msg_info(message string) {
  fmt.Println("[INFO]", message)
}

func download_file(filepath string, url string) (err error) {
  // Create the file
  out, err := os.Create(filepath)
  if err != nil  {
    return err
  }
  defer out.Close()
  // Get the data
  resp, err := http.Get(url)
  if err != nil {
    return err
  }
  defer resp.Body.Close()
  // Writer the body to file
  _, err = io.Copy(out, resp.Body)
  if err != nil  {
    return err
  }
  return nil
}

func hab_install(tree *toml.TomlTree) {
  if !tree.Has("habitat.location") {
    msg_error("[CONF] Group habitat it's required")
  }
  location := tree.Get("habitat.location").(string)
  if _, err := os.Stat(location); os.IsNotExist(err) {
    msg_info("Installing habitat at: "+location)
    // Install habitat
    if !tree.Has("habitat.download_url") {
      msg_error("[CONF] Habitat download URL it's required")
    }
    download_url := tree.Get("habitat.download_url").(string)
    os.MkdirAll(location,0755)
    download_file(location, download_url )
  }
}

func load_config() {
  // Declare variables
  // var project interface{}
  var toml_location string = "blue.toml"
  // Validate and Load TOML file
  if _, err := os.Stat(toml_location); os.IsNotExist(err) {
    msg_error("TOML file not found")
  }
  tree, err := toml.LoadFile(toml_location)
  if err != nil {
    msg_error(err.Error())
  } else {
    // Validate group project
    if !tree.Has("project") {
      msg_error("[CONF] Group project it's required")
    }
    g_project := tree.Get("project").(*toml.TomlTree)
    // Validate repository location
    if !g_project.Has("location") {
      msg_error("[CONF] Project's code location it's required")
    }
    location := g_project.Get("location").(string)
    if _, err := os.Stat(location); os.IsNotExist(err) {
      msg_error("[CONF] Project dir "+ location +" not found")
    }
    // Validate schemas
    if tree.Has("schema") {
      schemas := tree.Get("schema").(*toml.TomlTree)
      for _, schema_key := range schemas.Keys() {
        schema := schemas.Get(schema_key).(*toml.TomlTree)
        // Validate SQL dump
        if schema.Has("sql_dump") {
          sqldump := schema.Get("sql_dump").(string)
          if _, err := os.Stat(sqldump); os.IsNotExist(err) {
            msg_error("[CONF] File ("+ sqldump +") not found")
          }
        }
      }
    }
    // Validate stack
    if !tree.Has("stack") {
      msg_error("[CONF] Group stack it's required")
    }
    // Validate packages
    // TODO Validate if array it's empty
    if !tree.Has("stack.packages") {
      msg_error("[CONF] Stack's packages required")
    }
    rt := reflect.TypeOf(tree.Get("stack.packages"))
    if rt.Kind() != reflect.Slice {
      msg_error("[CONF] Stack's packages attribute should be an array")
    }
    // Load primary results to add extra slice (workaround)
    pres,_ := tree.Query("$.stack.packages[0:-1]")
    items := len(pres.Values())+1 // Workaround to get all the nodes
    results,_ := tree.Query("$.stack.packages[0:"+strconv.Itoa(items)+"]")
    // Validate habitat install
    hab_install(tree)
    for _,item := range results.Values() {
      fmt.Println(item)
    }
  }
}

func main() {
  load_config()
}

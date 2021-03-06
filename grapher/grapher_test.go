package lightwavegrapher

import (
  store "lightwavestore"
  "testing"
  "fmt"
  "log"
  "os"
  "time"
)

var schema = &Schema{ FileSchemas: map[string]*FileSchema {
    "application/x-test-file": &FileSchema{ EntitySchemas: map[string]*EntitySchema {
	"application/x-test-entity": &EntitySchema { FieldSchemas: map[string]*FieldSchema {
	    "text": &FieldSchema{ Type: TypeString, ElementType: TypeNone, Transformation: TransformationMerge } } } } } } }

type dummyTransformer struct {
  grapher *Grapher
}

func newDummyTransformer(grapher *Grapher) Transformer {
  t := &dummyTransformer{grapher: grapher}
  grapher.AddTransformer(t)
  return t
}

func (self *dummyTransformer) Kind() int {
  return TransformationMerge
}

func (self *dummyTransformer) DataType() int {
  return TypeString
}

// Interface towards the Grapher
func (self *dummyTransformer) TransformClientMutation(mutation MutationNode, rollback <-chan MutationNode) (err os.Error) {
  return
}

// Interface towards the Grapher
func (self *dummyTransformer) TransformMutation(mutation MutationNode, rollback <-chan MutationNode, concurrent []string) (err os.Error) {
  return
}

type dummyFederation struct {
}

func (self *dummyFederation) Forward(blobref string, users []string) {
  log.Printf("Forwarding %v to %v\n", blobref, users) 
}

func (self *dummyFederation) SetGrapher(idx *Grapher) {
}

func (self *dummyFederation) DownloadPermaNode(permission_blobref string) os.Error {
  return nil
}

func TestPermanode(t *testing.T) {
  s := store.NewSimpleBlobStore()
  sg := NewSimpleGraphStore()
  grapher := NewGrapher("a@b", schema, s, sg, &dummyFederation{})
  s.AddListener(grapher)
  
  blob1 := []byte(`{"type":"permanode", "signer":"a@b", "random":"perma1abc", "mimetype":"application/x-test-file"}`)
  blobref1 := store.NewBlobRef(blob1)
  blob2 := []byte(`{"type":"permanode", "signer":"a@b", "random":"perma2xyz", "mimetype":"application/x-test-file", "perma":"` + blobref1 + `"}`)
  blobref2 := store.NewBlobRef(blob2)
  
  s.StoreBlob(blob1, blobref1)
  s.StoreBlob(blob2, blobref2)
  
  time.Sleep(1000000000 * 2)

  perma, err := grapher.permaNode(blobref1)
  if perma == nil || err != nil {
    t.Fatal("Did not find perma node")
  }
  perma, err = grapher.permaNode(blobref2)
  if perma == nil || err != nil {
    t.Fatal("Did not find perma node")
  }
}

func TestPermanode2(t *testing.T) {
  s := store.NewSimpleBlobStore()
  sg := NewSimpleGraphStore()
  grapher := NewGrapher("a@b", schema, s, sg, &dummyFederation{})
  s.AddListener(grapher)

  blob1 := []byte(`{"type":"permanode", "signer":"a@b", "mimetype":"application/x-test-file", "random":"perma1abc"}`)
  blobref1 := store.NewBlobRef(blob1)
  blob2 := []byte(`{"type":"permanode", "signer":"a@b", "mimetype":"application/x-test-file", "random":"perma2xyz", "perma":"` + blobref1 + `"}`)
  blobref2 := store.NewBlobRef(blob2)

  // Insert them in the wrong order
  s.StoreBlob(blob2, blobref2)  
  s.StoreBlob(blob1, blobref1)
  
  time.Sleep(1000000000 * 2)

  perma, err := grapher.permaNode(blobref1)
  if perma == nil || err != nil {
    t.Fatal("Did not find perma node")
  }
  perma, err = grapher.permaNode(blobref2)
  if perma == nil || err != nil {
    t.Fatal("Did not find perma node")
  }
}

func TestPermanode3(t *testing.T) {
  fed := &dummyFederation{}
  s := store.NewSimpleBlobStore()
  sg := NewSimpleGraphStore()
  grapher := NewGrapher("a@b", schema, s, sg, fed)
  s.AddListener(grapher)
  newDummyTransformer(grapher)
  
  blob1 := []byte(`{"type":"permanode", "signer":"a@b", "mimetype":"application/x-test-file", "random":"perma1abc"}`)
  blobref1 := store.NewBlobRef(blob1)
  blob1b := []byte(`{"type":"keep", "signer":"a@b", "perma":"` + blobref1 + `"}`)
  blobref1b := store.NewBlobRef(blob1b)
  blob1c := []byte(`{"type":"entity", "signer":"a@b", "perma":"` + blobref1 + `", "mimetype": "application/x-test-entity", "content":"", "dep":["` + blobref1b + `"]}`)
  blobref1c := store.NewBlobRef(blob1c)
  blob2 := []byte(`{"type":"mutation", "signer":"a@b", "perma":"` + blobref1 + `", "dep":["` + blobref1c + `"], "op":{"$t":["Hello World"]}, "entity":"` + blobref1c + `", "field":"text"}`)
  blobref2 := store.NewBlobRef(blob2)
  blob3 := []byte(`{"type":"mutation", "signer":"a@b", "perma":"` + blobref1 + `", "dep":["` + blobref1c + `"], "op":{"$t":["Olla!!"]}, "entity":"` + blobref1c + `", "field":"text"}`)
  blobref3 := store.NewBlobRef(blob3)
  blob4 := []byte(`{"type":"mutation", "signer":"a@b", "perma":"` + blobref1 + `", "dep":["` + blobref2 + `"], "op":{"$t":[{"$s":11}, "??"]}, "entity":"` + blobref1c + `", "field":"text"}`)
  blobref4 := store.NewBlobRef(blob4)
  // Grant user foo@bar read access. At the same time this serves as an invitation
  blob5 := []byte(`{"type":"permission", "perma":"` + blobref1 + `", "signer":"a@b", "action":"invite", "dep":["` + blobref4 + `"], "user":"foo@bar", "allow":` + fmt.Sprintf("%v", Perm_Read) + `, "deny":0}`)
  blobref5 := store.NewBlobRef(blob5)
  // Fake a keep
  blob7 := []byte(`{"type":"keep", "signer":"foo@bar", "permission":"` + blobref5 + `", "perma":"` + blobref1 + `"}`)
  blobref7 := store.NewBlobRef(blob7)

  s.StoreBlob(blob1, blobref1)
  s.StoreBlob(blob1b, blobref1b)
  s.StoreBlob(blob1c, blobref1c)
  s.StoreBlob(blob2, blobref2)  
  s.StoreBlob(blob3, blobref3)  
  s.StoreBlob(blob4, blobref4)  
  s.StoreBlob(blob5, blobref5)
  s.StoreBlob(blob7, blobref7)
  
  time.Sleep(1000000000 * 2)

  perma, err := grapher.permaNode(blobref1)
  if perma == nil || err != nil {
    t.Fatal("Did not find perma node")
  }
  if !perma.hasKeep("a@b") {
    t.Fatal("Missing a keep for a@b")
  }
  if !perma.hasKeep("foo@bar") {
    t.Fatal("Missing a keep for foo@bar")
  }

  allow := perma.hasPermission("a@b", Perm_Read)
  if !allow {
    t.Fatal("Expected an allow for a@b")
  }

  allow = perma.hasPermission("x@y", Perm_Read)
  if allow {
    t.Fatal("Expected a deny")
  }

  allow = perma.hasPermission("foo@bar", Perm_Read)
  if !allow {
    t.Fatal("Expected an allow for foo@bar")
  }

  allow = perma.hasPermission("a@b", Perm_Invite | Perm_Expel)
  if !allow {
    t.Fatal("Expected an allow for Invite a@b")
  }

  users := perma.Followers()
  if len(users) != 2 {
    t.Fatalf("Wrong number of users: %v\n", users)
  }
  if (users[0] != "a@b" || users[1] != "foo@bar") && (users[1] != "a@b" || users[0] != "foo@bar") {
    t.Fatal("Wrong users")
  }

  users = perma.followersWithPermission(Perm_Read)
  if len(users) != 2 {
    t.Fatalf("Wrong number of users: %v\n", users)
  }
  if (users[0] != "a@b" || users[1] != "foo@bar") && (users[1] != "a@b" || users[0] != "foo@bar") {
    t.Fatal("Wrong users")
  }

  users = perma.followersWithPermission(Perm_Read | Perm_Invite)
  if len(users) != 1 {
    t.Fatalf("Wrong number of users: %v\n", users)
  }
  if users[0] != "a@b" {
    t.Fatal("Wrong users")
  }
}

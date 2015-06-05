package cfcomponent

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing with Ginkgo", func() {
	It("reading from json file", func() {
		file, err := ioutil.TempFile("", "config")
		defer func() {
			os.Remove(file.Name())
		}()
		Expect(err).NotTo(HaveOccurred())
		_, err = file.Write([]byte(`{"VarzUser":"User"}`))
		Expect(err).NotTo(HaveOccurred())

		err = file.Close()
		Expect(err).NotTo(HaveOccurred())

		config := &Config{}
		err = ReadConfigInto(config, file.Name())
		Expect(err).NotTo(HaveOccurred())

		Expect(config.VarzUser).To(Equal("User"))
	})

	It("reading my config from json file", func() {
		file, err := ioutil.TempFile("", "config")
		defer func() {
			os.Remove(file.Name())
		}()
		Expect(err).NotTo(HaveOccurred())
		_, err = file.Write([]byte(`{"VarzUser":"User", "CustomProperty":"CustomValue"}`))
		Expect(err).NotTo(HaveOccurred())

		err = file.Close()
		Expect(err).NotTo(HaveOccurred())

		config := &MyConfig{}
		err = ReadConfigInto(config, file.Name())
		Expect(err).NotTo(HaveOccurred())

		Expect(config.VarzUser).To(Equal("User"))
		Expect(config.CustomProperty).To(Equal("CustomValue"))
	})

	It("returns error if file not found", func() {
		config := &Config{}
		err := ReadConfigInto(config, "/foo/config.json")
		Expect(err).To(HaveOccurred())
	})

	It("returns error if invalid json", func() {
		file, err := ioutil.TempFile("", "config")
		defer func() {
			os.Remove(file.Name())
		}()
		Expect(err).NotTo(HaveOccurred())
		_, err = file.Write([]byte(`NotJson`))
		Expect(err).NotTo(HaveOccurred())

		err = file.Close()
		Expect(err).NotTo(HaveOccurred())

		config := &Config{}
		err = ReadConfigInto(config, file.Name())
		Expect(err).To(HaveOccurred())
	})
})

type MyConfig struct {
	Config
	CustomProperty string
}

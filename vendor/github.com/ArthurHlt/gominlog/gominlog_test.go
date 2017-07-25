package gominlog_test

import (
	. "github.com/ArthurHlt/gominlog"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"log"
)

var _ = Describe("Gominlog", func() {
	var buffer *gbytes.Buffer
	var logger *MinLog
	Context("When using classic minlog", func() {
		BeforeEach(func() {
			logger = NewClassicMinLog()
			buffer = gbytes.NewBuffer()
			logger.SetWriter(buffer)
		})
		It("should log inside stdout with flag log.Lshortfile | log.Ldate | log.Ltime and level to all", func() {
			logger.Debug("test debuglevel")
			Expect(buffer).Should(gbytes.Say("/gominlog.test/gominlog_test.go:[0-9]+ DEBUG: test debuglevel"))
			logger.Error("test %s", "errorlevel")
			Expect(buffer).Should(gbytes.Say("ERROR: test errorlevel"))
			logger.Warning("test warninglevel")
			Expect(buffer).Should(gbytes.Say("WARNING: test warninglevel"))
			logger.Severe("test severelevel")
			Expect(buffer).Should(gbytes.Say("SEVERE: test severelevel"))
			logger.Info("test infolevel")
			Expect(buffer).Should(gbytes.Say("INFO: test infolevel"))

			Expect(logger.GetLevel()).To(Equal(Lall))
			Expect(logger.IsColorized()).To(BeTrue())
			Expect(logger.GetLogger().Flags()).To(Equal(log.Lshortfile | log.Ldate | log.Ltime))
		})
		It("should give the file path by delimiting with packageName", func() {
			logger.SetPackageName("gominlog")
			logger.Debug("test debuglevel")
			Expect(buffer).Should(gbytes.Say("/gominlog/gominlog_test.go:[0-9]+ DEBUG: test debuglevel"))
		})
		It("should not show debug ouput when level set to info or more", func() {
			logger.SetLevel(Linfo)

			logger.Debug("test debuglevel")
			Expect(buffer).ShouldNot(gbytes.Say("DEBUG: test debuglevel"))
			logger.Error("test errorlevel")
			Expect(buffer).Should(gbytes.Say("ERROR: test errorlevel"))
			logger.Warning("test warninglevel")
			Expect(buffer).Should(gbytes.Say("WARNING: test warninglevel"))
			logger.Severe("test severelevel")
			Expect(buffer).Should(gbytes.Say("SEVERE: test severelevel"))
			logger.Info("test infolevel")
			Expect(buffer).Should(gbytes.Say("INFO: test infolevel"))
		})
		It("should not show debug and info ouput when level set to warning or more", func() {
			logger.SetLevel(Lwarning)

			logger.Debug("test debuglevel")
			Expect(buffer).ShouldNot(gbytes.Say("DEBUG: test debuglevel"))
			logger.Error("test errorlevel")
			Expect(buffer).Should(gbytes.Say("ERROR: test errorlevel"))
			logger.Warning("test warninglevel")
			Expect(buffer).Should(gbytes.Say("WARNING: test warninglevel"))
			logger.Severe("test severelevel")
			Expect(buffer).Should(gbytes.Say("SEVERE: test severelevel"))
			logger.Info("test infolevel")
			Expect(buffer).ShouldNot(gbytes.Say("INFO: test infolevel"))
		})
		It("should not show debug, info and warning ouput when level set to error or more", func() {
			logger.SetLevel(Lerror)

			logger.Debug("test debuglevel")
			Expect(buffer).ShouldNot(gbytes.Say("DEBUG: test debuglevel"))
			logger.Error("test errorlevel")
			Expect(buffer).Should(gbytes.Say("ERROR: test errorlevel"))
			logger.Warning("test warninglevel")
			Expect(buffer).ShouldNot(gbytes.Say("WARNING: test warninglevel"))
			logger.Severe("test severelevel")
			Expect(buffer).Should(gbytes.Say("SEVERE: test severelevel"))
			logger.Info("test infolevel")
			Expect(buffer).ShouldNot(gbytes.Say("INFO: test infolevel"))
		})
		It("should not show debug, info, warning and error ouput when level set to severe or more", func() {
			logger.SetLevel(Lsevere)

			logger.Debug("test debuglevel")
			Expect(buffer).ShouldNot(gbytes.Say("DEBUG: test debuglevel"))
			logger.Error("test errorlevel")
			Expect(buffer).ShouldNot(gbytes.Say("ERROR: test errorlevel"))
			logger.Warning("test warninglevel")
			Expect(buffer).ShouldNot(gbytes.Say("WARNING: test warninglevel"))
			logger.Severe("test severelevel")
			Expect(buffer).Should(gbytes.Say("SEVERE: test severelevel"))
			logger.Info("test infolevel")
			Expect(buffer).ShouldNot(gbytes.Say("INFO: test infolevel"))
		})
		It("should output nothing when level is off", func() {
			logger.SetLevel(Loff)

			logger.Debug("test debuglevel")
			Expect(buffer).ShouldNot(gbytes.Say("DEBUG: test debuglevel"))
			logger.Error("test errorlevel")
			Expect(buffer).ShouldNot(gbytes.Say("ERROR: test errorlevel"))
			logger.Warning("test warninglevel")
			Expect(buffer).ShouldNot(gbytes.Say("WARNING: test warninglevel"))
			logger.Severe("test severelevel")
			Expect(buffer).ShouldNot(gbytes.Say("SEVERE: test severelevel"))
			logger.Info("test infolevel")
			Expect(buffer).ShouldNot(gbytes.Say("INFO: test infolevel"))
		})
		AfterEach(func() {
			logger.SetLevel(Lall)
			logger.SetPackageName("")
		})
	})
	Context("when using a classic minlog with packageName given", func() {
		BeforeEach(func() {
			logger = NewClassicMinLogWithPackageName("gominlog")
			buffer = gbytes.NewBuffer()
			logger.SetWriter(buffer)
		})
		It("should give the file path by delimiting with packageName", func() {
			logger.Debug("test debuglevel")
			Expect(buffer).Should(gbytes.Say("/gominlog/gominlog_test.go:[0-9]+ DEBUG: test debuglevel"))

			Expect(logger.GetLevel()).To(Equal(Lall))
			Expect(logger.IsColorized()).To(BeTrue())
			Expect(logger.GetLogger().Flags()).To(Equal(log.Lshortfile | log.Ldate | log.Ltime))
		})
	})
	Context("when creating a minlog", func() {
		BeforeEach(func() {
			logger = NewMinLog("gominlog", Linfo, false, log.Llongfile | log.Ltime | log.Ldate)
			buffer = gbytes.NewBuffer()
			logger.SetWriter(buffer)
		})
		It("should give the file path by delimiting with packageName", func() {
			logger.Info("test infolevel")
			Expect(buffer).Should(gbytes.Say("github.com/ArthurHlt/gominlog/gominlog.go:[0-9]+: INFO: test infolevel"))

			logger.Debug("test debuglevel")
			Expect(buffer).ShouldNot(gbytes.Say("DEBUG: test debuglevel"))

			Expect(logger.GetLevel()).To(Equal(Linfo))
			Expect(logger.IsColorized()).To(BeFalse())
			Expect(logger.GetLogger().Flags()).To(Equal(log.Llongfile | log.Ltime | log.Ldate))
		})
	})
	Context("when creating a minlog with logger attach", func() {
		BeforeEach(func() {

			buffer = gbytes.NewBuffer()
			flags := log.Llongfile | log.Ltime | log.Ldate
			loggerLog := log.New(buffer, "", flags)
			logger = NewMinLogWithLogger("gominlog", Linfo, false, loggerLog)
		})
		It("should give the file path by delimiting with packageName", func() {
			logger.Info("test infolevel")
			Expect(buffer).Should(gbytes.Say("github.com/ArthurHlt/gominlog/gominlog.go:[0-9]+: INFO: test infolevel"))

			logger.Debug("test debuglevel")
			Expect(buffer).ShouldNot(gbytes.Say("DEBUG: test debuglevel"))

			Expect(logger.GetLevel()).To(Equal(Linfo))
			Expect(logger.GetLogger().Flags()).To(Equal(log.Llongfile | log.Ltime | log.Ldate))
		})
	})
})

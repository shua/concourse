package integration_test

import (
	"net/http"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/concourse/concourse/atc"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Fly CLI", func() {
	Describe("hide-pipeline", func() {
		Context("when the pipeline name is specified", func() {
			var (
				path string
				err  error
			)
			BeforeEach(func() {
				path, err = atc.CreatePathForRoute(atc.HidePipeline, map[string]string{"pipeline_name": "awesome-pipeline", "team_name": teamName})
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when the pipeline exists", func() {
				BeforeEach(func() {
					atcServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", path),
							ghttp.RespondWith(http.StatusOK, nil),
						),
					)
				})

				It("hides the pipeline", func() {
					Expect(func() {
						flyCmd := exec.Command(flyPath, "-t", targetName, "hide-pipeline", "-p", "awesome-pipeline")

						sess, err := gexec.Start(flyCmd, GinkgoWriter, GinkgoWriter)
						Expect(err).NotTo(HaveOccurred())

						Eventually(sess).Should(gbytes.Say(`hid 'awesome-pipeline'`))

						<-sess.Exited
						Expect(sess.ExitCode()).To(Equal(0))
					}).To(Change(func() int {
						return len(atcServer.ReceivedRequests())
					}).By(2))
				})
			})

			Context("when the pipeline doesn't exist", func() {
				BeforeEach(func() {
					atcServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", path),
							ghttp.RespondWith(http.StatusNotFound, nil),
						),
					)
				})

				It("prints helpful message", func() {
					Expect(func() {
						flyCmd := exec.Command(flyPath, "-t", targetName, "hide-pipeline", "-p", "awesome-pipeline")

						sess, err := gexec.Start(flyCmd, GinkgoWriter, GinkgoWriter)
						Expect(err).NotTo(HaveOccurred())

						Eventually(sess.Err).Should(gbytes.Say(`pipeline 'awesome-pipeline' not found`))

						<-sess.Exited
						Expect(sess.ExitCode()).To(Equal(1))
					}).To(Change(func() int {
						return len(atcServer.ReceivedRequests())
					}).By(2))
				})
			})
		})

		Context("when both the team and pipeline name are specified", func() {
			teamName := "other-team"
			var (
				path string
				err  error
			)
			BeforeEach(func() {
				path, err = atc.CreatePathForRoute(atc.HidePipeline, map[string]string{"pipeline_name": "awesome-pipeline", "team_name": teamName})
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when the pipeline exists", func() {
				BeforeEach(func() {
					atcServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v1/teams/"+teamName),
							ghttp.RespondWithJSONEncoded(http.StatusOK, atc.Team{
								Name: teamName,
							}),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", path),
							ghttp.RespondWith(http.StatusOK, nil),
						),
					)
				})

				It("hides the pipeline", func() {
					Expect(func() {
						flyCmd := exec.Command(flyPath, "-t", targetName, "hide-pipeline", "--team", teamName, "-p", "awesome-pipeline")

						sess, err := gexec.Start(flyCmd, GinkgoWriter, GinkgoWriter)
						Expect(err).NotTo(HaveOccurred())

						Eventually(sess).Should(gbytes.Say(`hid 'awesome-pipeline'`))

						<-sess.Exited
						Expect(sess.ExitCode()).To(Equal(0))
					}).To(Change(func() int {
						return len(atcServer.ReceivedRequests())
					}).By(3))
				})
			})

			Context("when the pipeline doesn't exist", func() {
				BeforeEach(func() {
					atcServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v1/teams/"+teamName),
							ghttp.RespondWithJSONEncoded(http.StatusOK, atc.Team{
								Name: teamName,
							}),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", path),
							ghttp.RespondWith(http.StatusNotFound, nil),
						),
					)
				})

				It("prints helpful message", func() {
					Expect(func() {
						flyCmd := exec.Command(flyPath, "-t", targetName, "hide-pipeline", "--team", teamName, "-p", "awesome-pipeline")

						sess, err := gexec.Start(flyCmd, GinkgoWriter, GinkgoWriter)
						Expect(err).NotTo(HaveOccurred())

						Eventually(sess.Err).Should(gbytes.Say(`pipeline 'awesome-pipeline' not found`))

						<-sess.Exited
						Expect(sess.ExitCode()).To(Equal(1))
					}).To(Change(func() int {
						return len(atcServer.ReceivedRequests())
					}).By(3))
				})
			})
		})

		Context("when the pipline name is not specified", func() {
			It("errors", func() {
				Expect(func() {
					flyCmd := exec.Command(flyPath, "-t", targetName, "hide-pipeline")

					sess, err := gexec.Start(flyCmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())

					<-sess.Exited
					Expect(sess.ExitCode()).To(Equal(1))
				}).To(Change(func() int {
					return len(atcServer.ReceivedRequests())
				}).By(0))
			})
		})

		Context("when specifying a pipeline name with a '/' character in it", func() {
			It("fails and says '/' characters are not allowed", func() {
				flyCmd := exec.Command(flyPath, "-t", targetName, "hide-pipeline", "-p", "forbidden/pipelinename")

				sess, err := gexec.Start(flyCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				<-sess.Exited
				Expect(sess.ExitCode()).To(Equal(1))

				Expect(sess.Err).To(gbytes.Say("error: pipeline name cannot contain '/'"))
			})
		})

	})
})

package authentication

const emailTemplate = `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml" lang="en" xml:lang="en">
  <head>
      <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
      <meta name="viewport" content="width=device-width" />
      <meta name="robots" content="noindex" />
      <title>{{.Title}}</title>
      <style>
          /**
    * IMPORTANT:
    * Please read before changing anything, CSS involved in our HTML emails is
    * extremely specific and written a certain way for a reason. It might not make
    * sense in a normal setting but Outlook loves it this way.
    *
    * !!! div[style*="margin: 16px 0"] allows us to target a weird margin
    * !!! bug in Android's email client.
    * !!! Do not remove.
    *
    * Also, the img files are hosted on S3. Please don't break these URLs!
    * The images are also versioned by date, so please update the URLs accordingly
    * if you create new versions
    *
  ***/

          /**
    * # Root
    * - CSS resets and general styles go here.
  **/

          html,
          body,
          a,
          span,
          div[style*="margin: 16px 0"] {
              border: 0 !important;
              margin: 0 !important;
              outline: 0 !important;
              padding: 0 !important;
              text-decoration: none !important;
          }

          a,
          span,
          td,
          th {
              -webkit-font-smoothing: antialiased !important;
              -moz-osx-font-smoothing: grayscale !important;
          }

          /**
    * # Delink
    * - Classes for overriding clients which creates links out of things like
    *   emails, addresses, phone numbers, etc.
  **/

          span.st-Delink a {
              color: #414552 !important;
              text-decoration: none !important;
          }

          /** Modifier: preheader */
          span.st-Delink.st-Delink--preheader a {
              color: #ffffff !important;
              text-decoration: none !important;
          }
          /** */

          /** Modifier: title */
          span.st-Delink.st-Delink--title a {
              color: #414552 !important;
              text-decoration: none !important;
          }
          /** */

          /** Modifier: footer */
          span.st-Delink.st-Delink--footer a {
              color: #687385 !important;
              text-decoration: none !important;
          }
          /** */

          /**
    * # Header
  **/

          table.st-Header td.st-Header-background div.st-Header-area {
              height: 76px !important;
              width: 600px !important;
              background-repeat: no-repeat !important;
              background-size: 600px 76px !important;
          }

          table.st-Header td.st-Header-logo div.st-Header-area {
              height: 29px !important;
              width: 70px !important;
              background-repeat: no-repeat !important;
              background-size: 70px 29px !important;
          }

          table.st-Header
              td.st-Header-logo.st-Header-logo--atlasAzlo
              div.st-Header-area {
              height: 21px !important;
              width: 216px !important;
              background-repeat: no-repeat !important;
              background-size: 216px 21px !important;
          }

          /**
    * # Retina
    * - Targets high density displays and devices smaller than 768px.
    *
    * ! For mobile specific styling, see # Mobile.
  **/

          @media (-webkit-min-device-pixel-ratio: 1.25),
              (min-resolution: 120dpi),
              all and (max-width: 768px) {
              /**
      * # Target
      * - Hides images in these devices to display the larger version as a
      *   background image instead.
    **/

              /** Modifier: mobile */
              div.st-Target.st-Target--mobile img {
                  display: none !important;
                  margin: 0 !important;
                  max-height: 0 !important;
                  min-height: 0 !important;
                  mso-hide: all !important;
                  padding: 0 !important;
                  font-size: 0 !important;
                  line-height: 0 !important;
              }
              /** */

              /**
      * # Header
    **/

              table.st-Header td.st-Header-background div.st-Header-area {
                  background-image: url("https://stripe-images.s3.amazonaws.com/html_emails/2017-08-21/header/Header-background.png") !important;
              }

              /** Modifier: white */
              table.st-Header.st-Header--white
                  td.st-Header-background
                  div.st-Header-area {
                  background-image: url("https://stripe-images.s3.amazonaws.com/html_emails/2017-08-21/header/Header-background--white.png") !important;
              }
              /** */

              /** Modifier: simplified */
              table.st-Header.st-Header--simplified
                  td.st-Header-logo
                  div.st-Header-area {
                  background-image: url("https://stripe-images.s3.amazonaws.com/html_emails/2023-03-30/header/stripe_logo_blurple_email_highres.png") !important;
              }
              /** */

              /** Modifier: simplified + atlasAzlo */
              table.st-Header.st-Header--simplified
                  td.st-Header-logo.st-Header-logo--atlasAzlo
                  div.st-Header-area {
                  background-image: url("https://stripe-images.s3.amazonaws.com/html_emails/2018-05-02/header/Header-logo--atlasAzlo.png") !important;
              }
              /** */
          }

          /**
    * # Mobile
    * - This affects emails views in clients less than 600px wide.
  **/

          @media only screen and (max-width: 600px) {
              /**
      * # Wrapper
    **/

              table.st-Wrapper,
              table.st-Width.st-Width--mobile {
                  min-width: 100% !important;
                  width: 100% !important;
                  border-radius: 0px !important;
              }

              /**
      * # Spacer
    **/

              /** Modifier: gutter */
              td.st-Spacer.st-Spacer--gutter {
                  width: 16px !important;
              }
              /** */

              /** Modifier: td.kill */
              td.st-Spacer.st-Spacer--kill {
                  width: 0 !important;
              }
              td.st-Spacer.st-Spacer--height {
                  height: 0 !important;
              }
              /** */

              /** Modifier: div.kill */
              div.st-Spacer.st-Spacer--kill {
                  height: 0px !important;
              }
              /** */

              /** Modifier: footer */
              .st-Mobile--footer {
                  text-align: left !important;
              }
              /** */

              /**
      * # Font
    **/

              /** Modifier: title */
              td.st-Font.st-Font--title,
              td.st-Font.st-Font--title span,
              td.st-Font.st-Font--title a {
                  font-size: 20px !important;
                  line-height: 28px !important;
                  font-weight: 700 !important;
              }
              /** */

              /** Modifier: header */
              td.st-Font.st-Font--header,
              td.st-Font.st-Font--header span,
              td.st-Font.st-Font--header a {
                  font-size: 16px !important;
                  line-height: 24px !important;
              }
              /** */

              /** Modifier: body */
              td.st-Font.st-Font--body,
              td.st-Font.st-Font--body span,
              td.st-Font.st-Font--body a {
                  font-size: 16px !important;
                  line-height: 24px !important;
              }
              /** */

              /** Modifier: caption */
              td.st-Font.st-Font--caption,
              td.st-Font.st-Font--caption span,
              td.st-Font.st-Font--caption a {
                  font-size: 12px !important;
                  line-height: 16px !important;
              }
              /** */

              /**
      * # Header
    **/
              table.st-Header td.st-Header-background div.st-Header-area {
                  margin: 0 !important;
                  width: auto !important;
                  background-position: 0 0 !important;
              }

              table.st-Header td.st-Header-background div.st-Header-area {
                  background-image: url("https://stripe-images.s3.amazonaws.com/html_emails/2017-08-21/header/Header-background--mobile.png") !important;
              }

              /** Modifier: white */
              table.st-Header.st-Header--white
                  td.st-Header-background
                  div.st-Header-area {
                  background-image: url("https://stripe-images.s3.amazonaws.com/html_emails/2017-08-21/header/Header-background--white--mobile.png") !important;
              }
              /** */

              /** Modifier: simplified */
              table.st-Header.st-Header--simplified td.st-Header-logo {
                  width: auto !important;
              }

              table.st-Header.st-Header--simplified td.st-Header-spacing {
                  width: 0 !important;
              }

              table.st-Header.st-Header--simplified
                  td.st-Header-logo
                  div.st-Header-area {
                  margin: 0 !important;
                  background-image: url("https://stripe-images.s3.amazonaws.com/html_emails/2023-03-30/header/stripe_logo_blurple_email_highres.png") !important;
              }

              table.st-Header.st-Header--simplified
                  td.st-Header-logo.st-Header-logo--atlasAzlo
                  div.st-Header-area {
                  margin: 0 auto !important;
                  background-image: url("https://stripe-images.s3.amazonaws.com/html_emails/2018-05-02/header/Header-logo--atlasAzlo.png") !important;
              }
              /** */

              /**
      * # Divider
    **/

              table.st-Divider td.st-Spacer.st-Spacer--gutter,
              tr.st-Divider td.st-Spacer.st-Spacer--gutter {
                  background-color: #e6ebf1;
              }

              /**
      * # Blocks
    **/

              table.st-Blocks table.st-Blocks-inner {
                  border-radius: 0 !important;
              }

              table.st-Blocks
                  table.st-Blocks-inner
                  table.st-Blocks-item
                  td.st-Blocks-item-cell {
                  display: block !important;
              }

              /**
      * # Hero Units
    **/

              /* Hides dividers in hero units so that vertical spacing remains consistent */
              table.st-Hero-Container td.st-Spacer--divider {
                  display: none !important;
                  margin: 0 !important;
                  max-height: 0 !important;
                  min-height: 0 !important;
                  mso-hide: all !important;
                  padding: 0 !important;

                  font-size: 0 !important;
                  line-height: 0 !important;
              }

              table.st-Hero-Responsive {
                  margin: 16px auto !important;
              }

              /**
      * # Button
    **/

              table.st-Button td.st-Button-area,
              table.st-Button td.st-Button-area a.st-Button-link,
              table.st-Button td.st-Button-area span.st-Button-internal {
                  height: 40px !important;
                  line-height: 24px !important;
                  font-size: 16px !important;
              }
          }

          @media (-webkit-min-device-pixel-ratio: 1.25),
              (min-resolution: 120dpi),
              all and (max-width: 768px) {
              /**
      * # mobile image
     **/
              div.st-Target.st-Target--mobile img {
                  display: none !important;
                  margin: 0 !important;
                  max-height: 0 !important;
                  min-height: 0 !important;
                  mso-hide: all !important;
                  padding: 0 !important;

                  font-size: 0 !important;
                  line-height: 0 !important;
              }

              /**
      * # document-list-item image
     **/
              div.st-Icon.st-Icon--document {
                  background-image: url("https://stripe-images.s3.amazonaws.com/notifications/icons/document--16--regular.png") !important;
              }
          }
      </style>
  </head>
  <body
      class="st-Email"
      bgcolor="#f6f9fc"
      style="
          border: 0;
          margin: 0;
          padding: 0;
          -webkit-text-size-adjust: 100%;
          -ms-text-size-adjust: 100%;
          min-width: 100%;
          width: 100%;
      "
      override="fix"
  >
      <!-- Background -->
      <table
          class="st-Background"
          bgcolor="#f6f9fc"
          border="0"
          cellpadding="0"
          cellspacing="0"
          width="100%"
          style="border: 0; margin: 0; padding: 0"
      >
          <tbody>
              <tr>
                  <td
                      class="st-Spacer st-Spacer--kill st-Spacer--height"
                      height="64"
                  >
                      <div class="st-Spacer st-Spacer--kill">&nbsp;</div>
                  </td>
              </tr>
              <tr>
                  <td style="border: 0; margin: 0; padding: 0">
                      <!-- Wrapper -->
                      <table
                          class="st-Wrapper"
                          align="center"
                          bgcolor="#ffffff"
                          border="0"
                          cellpadding="0"
                          cellspacing="0"
                          width="600"
                          style="
                              border-top-left-radius: 16px;
                              border-top-right-radius: 16px;
                              margin: 0 auto;
                              min-width: 600px;
                          "
                      >
                          <tbody>
                              <tr>
                                  <td
                                      style="border: 0; margin: 0; padding: 0"
                                  >
                                      <!-- Header -->
                                      <div
                                          style="
                                              background-color: #f6f9fc;
                                              padding-top: 20px;
                                          "
                                      >
                                          <table
                                              dir="ltr"
                                              class="Section Header"
                                              width="100%"
                                              style="
                                                  border: 0;
                                                  border-collapse: collapse;
                                                  margin: 0;
                                                  padding: 0;
                                                  background-color: #ffffff;
                                              "
                                          >
                                              <tbody>
                                                  <tr>
                                                      <td
                                                          class="Header-left Target"
                                                          style="
                                                              background-color: #fb4240;
                                                              border: 0;
                                                              border-collapse: collapse;
                                                              margin: 0;
                                                              padding: 0;
                                                              -webkit-font-smoothing: antialiased;
                                                              -moz-osx-font-smoothing: grayscale;
                                                              font-size: 0;
                                                              line-height: 0px;
                                                              mso-line-height-rule: exactly;
                                                              background-size: 100%
                                                                  100%;
                                                              border-top-left-radius: 5px;
                                                          "
                                                          align="right"
                                                          height="156"
                                                          valign="bottom"
                                                          width="252"
                                                      >
                                                          <a
                                                              href="http://base.al"
                                                              target="_blank"
                                                              style="
                                                                  -webkit-font-smoothing: antialiased;
                                                                  -moz-osx-font-smoothing: grayscale;
                                                                  outline: 0;
                                                                  text-decoration: none;
                                                              "
                                                          >
                                                              <img
                                                                  alt=""
                                                                  height="156"
                                                                  width="252"
                                                                  src="https://stripe-images.s3.amazonaws.com/notifications/hosted/20180110/Header/Left.png"
                                                                  style="
                                                                      display: block;
                                                                      border: 0;
                                                                      line-height: 100%;
                                                                      width: 100%;
                                                                  "
                                                              />
                                                          </a>
                                                      </td>
                                                      <td
                                                          class="Header-icon Target"
                                                          style="
                                                              background-color: #fb4240;
                                                              border: 0;
                                                              border-collapse: collapse;
                                                              margin: 0;
                                                              padding: 0;
                                                              -webkit-font-smoothing: antialiased;
                                                              -moz-osx-font-smoothing: grayscale;
                                                              font-size: 0;
                                                              line-height: 0px;
                                                              mso-line-height-rule: exactly;
                                                              background-size: 100%
                                                                  100%;
                                                              width: 96px !important;
                                                          "
                                                          align="center"
                                                          height="156"
                                                          valign="bottom"
                                                      >
                                                          <a
                                                              href="http://base.al"
                                                              target="_blank"
                                                              style="
                                                                  -webkit-font-smoothing: antialiased;
                                                                  -moz-osx-font-smoothing: grayscale;
                                                                  outline: 0;
                                                                  text-decoration: none;
                                                              "
                                                          >
                                                              <img
                                                                  alt=""
                                                                  height="156"
                                                                  width="96"
                                                                  src="https://stripe-images.s3.amazonaws.com/emails/acct_1PMXIdP5EhigQEI7/1/twelve_degree_icon@2x.png"
                                                                  style="
                                                                      display: block;
                                                                      border: 0;
                                                                      line-height: 100%;
                                                                  "
                                                              />
                                                          </a>
                                                      </td>
                                                      <td
                                                          class="Header-right Target"
                                                          style="
                                                              background-color: #fb4240;
                                                              border: 0;
                                                              border-collapse: collapse;
                                                              margin: 0;
                                                              padding: 0;
                                                              -webkit-font-smoothing: antialiased;
                                                              -moz-osx-font-smoothing: grayscale;
                                                              font-size: 0;
                                                              line-height: 0px;
                                                              mso-line-height-rule: exactly;
                                                              background-size: 100%
                                                                  100%;
                                                              border-top-right-radius: 5px;
                                                          "
                                                          align="left"
                                                          height="156"
                                                          valign="bottom"
                                                          width="252"
                                                      >
                                                          <a
                                                              href="http://base.al"
                                                              target="_blank"
                                                              style="
                                                                  -webkit-font-smoothing: antialiased;
                                                                  -moz-osx-font-smoothing: grayscale;
                                                                  outline: 0;
                                                                  text-decoration: none;
                                                              "
                                                          >
                                                              <img
                                                                  alt=""
                                                                  height="156"
                                                                  width="252"
                                                                  src="https://stripe-images.s3.amazonaws.com/notifications/hosted/20180110/Header/Right.png"
                                                                  style="
                                                                      display: block;
                                                                      border: 0;
                                                                      line-height: 100%;
                                                                      width: 100%;
                                                                  "
                                                              />
                                                          </a>
                                                      </td>
                                                  </tr>
                                              </tbody>
                                          </table>
                                      </div>
                                      <!-- Email Title -->
                                      <table
                                          class="st-Copy st-Copy--caption st-Width st-Width--mobile"
                                          border="0"
                                          cellpadding="0"
                                          cellspacing="0"
                                          width="600"
                                          style="min-width: 600px"
                                      >
                                          <tbody>
                                              <tr>
                                                  <td
                                                      class="Content Title-copy Font Font--title"
                                                      align="center"
                                                      style="
                                                          border: 0;
                                                          border-collapse: collapse;
                                                          margin: 0;
                                                          padding: 0;
                                                          -webkit-font-smoothing: antialiased;
                                                          -moz-osx-font-smoothing: grayscale;
                                                          width: 472px;
                                                          font-family: -apple-system,
                                                              BlinkMacSystemFont,
                                                              &quot;Segoe UI&quot;,
                                                              Roboto,
                                                              &quot;Helvetica Neue&quot;,
                                                              Ubuntu,
                                                              sans-serif;
                                                          mso-line-height-rule: exactly;
                                                          vertical-align: middle;
                                                          color: #32325d;
                                                          font-size: 24px;
                                                          line-height: 32px;
                                                      "
                                                  >
                                                      {{.Title}}
                                                  </td>
                                              </tr>
                                              <tr>
                                                  <td
                                                      class="st-Spacer st-Spacer--stacked"
                                                      colspan="3"
                                                      height="12"
                                                      style="
                                                          border: 0;
                                                          margin: 0;
                                                          padding: 0;
                                                          font-size: 1px;
                                                          line-height: 1px;
                                                          mso-line-height-rule: exactly;
                                                      "
                                                  >
                                                      <div
                                                          class="st-Spacer st-Spacer--filler"
                                                      ></div>
                                                  </td>
                                              </tr>
                                          </tbody>
                                      </table>
                                      <!-- Main Content -->
                                      <table
                                          class="st-Copy st-Width st-Width--mobile"
                                          border="0"
                                          cellpadding="0"
                                          cellspacing="0"
                                          width="600"
                                          style="min-width: 600px"
                                      >
                                          <tbody>
                                              <tr>
                                                  <td
                                                      class="st-Spacer st-Spacer--gutter"
                                                      style="
                                                          border: 0;
                                                          margin: 0;
                                                          padding: 0;
                                                          font-size: 1px;
                                                          line-height: 1px;
                                                          mso-line-height-rule: exactly;
                                                      "
                                                      width="48"
                                                  >
                                                      <div
                                                          class="st-Spacer st-Spacer--filler"
                                                      ></div>
                                                  </td>
                                                  <td
                                                      style="
                                                          border: 0;
                                                          margin: 0;
                                                          padding: 0;
                                                          color: #414552 !important;
                                                          font-family: -apple-system,
                                                              &quot;SF Pro Display&quot;,
                                                              &quot;SF Pro Text&quot;,
                                                              &quot;Helvetica&quot;,
                                                              sans-serif;
                                                          font-weight: 400;
                                                          font-size: 16px;
                                                          line-height: 24px;
                                                      "
                                                  >
                                                      {{.Content}}
                                                  </td>
                                                  <td
                                                      class="st-Spacer st-Spacer--gutter"
                                                      style="
                                                          border: 0;
                                                          margin: 0;
                                                          padding: 0;
                                                          font-size: 1px;
                                                          line-height: 1px;
                                                          mso-line-height-rule: exactly;
                                                      "
                                                      width="48"
                                                  >
                                                      <div
                                                          class="st-Spacer st-Spacer--filler"
                                                      ></div>
                                                  </td>
                                              </tr>
                                          </tbody>
                                      </table>
                                      <!-- Footer -->
                                      <table
                                          class="st-Copy st-Width st-Width--mobile"
                                          border="0"
                                          cellpadding="0"
                                          cellspacing="0"
                                          width="600"
                                          style="min-width: 600px"
                                      >
                                          <tbody>
                                              <tr>
                                                  <td
                                                      class="st-Spacer st-Spacer--divider"
                                                      colspan="3"
                                                      height="20"
                                                      style="
                                                          border: 0;
                                                          margin: 0;
                                                          padding: 0;
                                                          font-size: 1px;
                                                          line-height: 1px;
                                                          max-height: 1px;
                                                          mso-line-height-rule: exactly;
                                                      "
                                                  >
                                                      <div
                                                          class="st-Spacer st-Spacer--filler"
                                                      ></div>
                                                  </td>
                                              </tr>
                                              <tr>
                                                  <td
                                                      class="st-Spacer st-Spacer--gutter"
                                                      style="
                                                          border: 0;
                                                          margin: 0;
                                                          padding: 0;
                                                          font-size: 1px;
                                                          line-height: 1px;
                                                          mso-line-height-rule: exactly;
                                                      "
                                                      width="48"
                                                  >
                                                      <div
                                                          class="st-Spacer st-Spacer--filler"
                                                      ></div>
                                                  </td>
                                                  <td
                                                      style="
                                                          border: 0;
                                                          margin: 0;
                                                          padding: 0;
                                                          color: #8898aa;
                                                          font-family: -apple-system,
                                                              &quot;SF Pro Display&quot;,
                                                              &quot;SF Pro Text&quot;,
                                                              &quot;Helvetica&quot;,
                                                              sans-serif;
                                                          font-weight: 400;
                                                          font-size: 12px;
                                                          line-height: 16px;
                                                      "
                                                  >
                                                      If you have any
                                                      questions, contact us at
                                                      <a
                                                          style="
                                                              border: 0;
                                                              margin: 0;
                                                              padding: 0;
                                                              color: #625afa !important;
                                                              font-weight: bold;
                                                              text-decoration: none;
                                                          "
                                                          href="mailto:support@base.al"
                                                          >support@base.al</a
                                                      >.
                                                  </td>
                                                  <td
                                                      class="st-Spacer st-Spacer--gutter"
                                                      style="
                                                          border: 0;
                                                          margin: 0;
                                                          padding: 0;
                                                          font-size: 1px;
                                                          line-height: 1px;
                                                          mso-line-height-rule: exactly;
                                                      "
                                                      width="48"
                                                  >
                                                      <div
                                                          class="st-Spacer st-Spacer--filler"
                                                      ></div>
                                                  </td>
                                              </tr>
                                              <tr>
                                                  <td
                                                      class="st-Spacer st-Spacer--divider"
                                                      colspan="3"
                                                      height="20"
                                                      style="
                                                          border: 0;
                                                          margin: 0;
                                                          padding: 0;
                                                          font-size: 1px;
                                                          line-height: 1px;
                                                          max-height: 1px;
                                                          mso-line-height-rule: exactly;
                                                      "
                                                  >
                                                      <div
                                                          class="st-Spacer st-Spacer--filler"
                                                      ></div>
                                                  </td>
                                              </tr>
                                          </tbody>
                                      </table>
                                  </td>
                              </tr>
                          </tbody>
                      </table>
                      <!-- /Wrapper -->
                  </td>
              </tr>
          </tbody>
      </table>
      <!-- /Background -->
  </body>
</html>
`

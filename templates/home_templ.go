// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.543
package templates

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

func Home() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)
		if !templ_7745c5c3_IsBuffer {
			templ_7745c5c3_Buffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<main style=\"display:flex; flex-direction:column; align-items:center;\"><h1>Enter your name</h1><div style=\"width:fit;\"><input type=\"text\" style=\"padding:1rem; margin:0.5rem; font-size:1.25rem;\" id=\"name\" placeholder=\"Name\"><div onclick=\"start()\" style=\"user-select:none; cursor:pointer; display:flex; flex-direction:column; justify-content:center; text-align:center; font-size:2rem; margin:0.5rem; border-radius:0.5rem; background-color:palegreen;\">→</div></div><script>\r\n            function start() {\r\n                var name = document.getElementById('name').value;\r\n                if (name) {\r\n                    window.location.href = `/play/${name}`;\r\n                }\r\n            }\r\n        </script></main>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}

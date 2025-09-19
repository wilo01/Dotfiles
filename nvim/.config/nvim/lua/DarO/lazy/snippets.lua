return {
   {
      "hrsh7th/nvim-cmp",
      event = "InsertEnter",
      dependencies = {
         { "hrsh7th/cmp-buffer" },
         { "hrsh7th/cmp-path" },
         { "hrsh7th/cmp-nvim-lsp" },
      },
      config = function()
         local dynamic_cache = {}
         local CACHE_TTL = 5000

         local function cached_dynamic(fn, cache_key)
            return function()
               local now = vim.uv.now()
               local cached = dynamic_cache[cache_key]

               if cached and (now - cached.time) < CACHE_TTL then
                  return cached.value
               end

               local ok, result = pcall(fn)
               local value = ok and result or ""
               dynamic_cache[cache_key] = { value = value, time = now }
               return value
            end
         end

         local get_date = cached_dynamic(function()
            return os.date("%d.%m.%Y")
         end, "date")

         local get_timestamp = cached_dynamic(function()
            return os.date("%d.%m.%Y %H:%M:%S")
         end, "timestamp")

         local get_filename = cached_dynamic(function()
            return vim.fn.expand("%:p")
         end, "filename")

         local get_pwd = cached_dynamic(function()
            return vim.fn.shellescape(vim.fn.getcwd())
         end, "pwd")

         local registry = {
            all = {},
            javascript = {},
            typescript = {},
            javascriptreact = {},
            typescriptreact = {},
            sh = {},
            python = {},
            lua = {},
            markdown = {},
            text = {},
            sql = {},
            plsql = {},
            apex = {},
            xml = {},
            go = {},
            rust = {},
            html = {},
            css = {},
            scss = {},
            yaml = {},
            dockerfile = {},
            json = {},
         }

         local function snip(filetype, trigger, body, description)
            if not registry[filetype] then
               registry[filetype] = {}
            end
            registry[filetype][trigger] = {
               trigger = trigger,
               body = body,
               description = description or trigger,
            }
         end

         local code_languages = {
            javascript = "javascript",
            js = "javascript",
            typescript = "typescript",
            ts = "typescript",
            html = "html",
            css = "css",
            scss = "scss",
            sass = "sass",
            jsx = "jsx",
            tsx = "tsx",
            python = "python",
            py = "python",
            java = "java",
            go = "go",
            rust = "rust",
            rs = "rust",
            c = "c",
            cpp = "cpp",
            csharp = "csharp",
            cs = "csharp",
            ruby = "ruby",
            rb = "ruby",
            php = "php",
            kotlin = "kotlin",
            swift = "swift",
            scala = "scala",
            sql = "sql",
            plsql = "plsql",
            mysql = "mysql",
            postgresql = "postgresql",
            postgres = "postgresql",
            bash = "bash",
            sh = "bash",
            zsh = "zsh",
            powershell = "powershell",
            ps1 = "powershell",
            lua = "lua",
            vim = "vim",
            perl = "perl",
            json = "json",
            yml = "yaml",
            xml = "xml",
            toml = "toml",
            ini = "ini",
            env = "env",
            dockerfile = "dockerfile",
            docker = "dockerfile",
            nginx = "nginx",
            apache = "apache",
            terraform = "terraform",
            tf = "terraform",
            markdown = "markdown",
            md = "markdown",
            latex = "latex",
            tex = "latex",
            apex = "apex",
            oracle = "sql",
            r = "r",
            matlab = "matlab",
            julia = "julia",
            haskell = "haskell",
            elixir = "elixir",
            erlang = "erlang",
            clojure = "clojure",
            dart = "dart",
         }

         local function setup_snippets()
            snip("all", "todo", "[ ] TODO: $0")
            snip("all", "todo-", "- [ ] TODO: $0")
            snip("all", "todo--", "-- [ ] TODO: $0")
            snip("all", "datetime", get_timestamp, "Insert current date and time")
            snip("all", "timestamp", get_timestamp, "Insert current timestamp")
            snip("all", "date", get_date, "Insert current date")
            snip("all", "pwd", get_pwd, "Insert current working directory")
            snip("all", "filename", get_filename, "Insert current filename with path")

            local sql_filetypes = { "sql", "plsql", "apex", "xml" }
            for _, ft in ipairs(sql_filetypes) do
               snip(ft, "select", "select ${1:*}\nfrom ${2:table_name}\nwhere ${3:condition};$0")
               snip(ft, "selectjoin",
                  "select ${1:t1.*}, ${2:t2.*}\nfrom ${3:table1} t1\ninner join ${4:table2} t2 on t1.${5:id} = t2.${5:id}\nwhere ${6:condition};$0")
               snip(ft, "selectcte",
                  "with ${1:cte_name} as (\n\tselect ${2:*}\n\tfrom ${3:table}\n\twhere ${4:condition}\n)\nselect * from ${1:cte_name};$0")
               snip(ft, "insert", "insert into ${1:table} (${2:columns})\nvalues (${3:values});$0")
               snip(ft, "update", "update ${1:table}\nset ${2:column} = ${3:value}\nwhere ${4:condition};$0")
               snip(ft, "delete", "delete from ${1:table}\nwhere ${2:condition};$0")
               snip(ft, "merge",
                  "merge into ${1:target_table} t\nusing ${2:source_table} s\non (t.${3:id} = s.${3:id})\nwhen matched then\n\tupdate set t.${4:col} = s.${4:col}\nwhen not matched then\n\tinsert (${5:columns})\n\tvalues (${6:values});$0")

               snip(ft, "createtable",
                  "create table ${1:table_name} (\n\t${2:id} number,\n\t${3:column_name} varchar2(${4:100}),\n\tconstraint ${1:table_name}_pk primary key (${2:id})\n);$0")
               snip(ft, "createindex", "create index ${1:idx_name} on ${2:table}(${3:column});$0")
               snip(ft, "createsequence",
                  "create sequence ${1:seq_name}\nminvalue 1\nstart with ${2:1}\nincrement by ${3:1}\nnocache order;$0")
               snip(ft, "createview",
                  "create or replace view ${1:view_name} (\n\t${2:column1},\n\t${3:column2}\n) as\nselect ${4:*}\nfrom ${5:table}\nwhere ${6:condition};$0")
               snip(ft, "altertable", "alter table ${1:table_name}\n${2:add} ${3:column_name} ${4:datatype};$0")
               snip(ft, "droptable", "drop table ${1:table_name} cascade constraints;$0")
               snip(ft, "truncate", "truncate table ${1:table_name};$0")
               snip(ft, "grant", "grant ${1:select, insert, update, delete} on ${2:table} to ${3:user};$0")
               snip(ft, "revoke", "revoke ${1:all} on ${2:table} from ${3:user};$0")
               snip(ft, "synonym", "create or replace synonym ${1:synonym_name} for ${2:schema}.${3:object};$0")

               snip(ft, "commit", "commit;$0")
               snip(ft, "rollback", "rollback;$0")
               snip(ft, "savepoint", "savepoint ${1:savepoint_name};$0")

               snip(ft, "exists", "exists (select 1 from ${1:table} where ${2:condition})$0")
               snip(ft, "case",
                  "case\n\twhen ${1:condition} then ${2:result}\n\twhen ${3:condition} then ${4:result}\n\telse ${5:default}\nend$0")
               snip(ft, "decode", "decode(${1:expr}, ${2:search}, ${3:result}, ${4:default})$0")
               snip(ft, "nvl", "nvl(${1:value}, ${2:default})$0")
               snip(ft, "nvl2", "nvl2(${1:value}, ${2:not_null}, ${3:null})$0")
               snip(ft, "coalesce", "coalesce(${1:value1}, ${2:value2}, ${3:value3})$0")

               snip(ft, "rownum", "where rownum <= ${1:10}$0")
               snip(ft, "analyticfn", "row_number() over (partition by ${1:column} order by ${2:column})$0")
               snip(ft, "rank", "rank() over (partition by ${1:column} order by ${2:column})$0")
               snip(ft, "denserank", "dense_rank() over (partition by ${1:column} order by ${2:column})$0")
               snip(ft, "leadlag", "lead(${1:column}, ${2:1}) over (order by ${3:column})$0")
               snip(ft, "pivot",
                  "select * from (\n\tselect ${1:columns}\n\tfrom ${2:table}\n)\npivot (\n\t${3:max}(${4:value_column})\n\tfor ${5:pivot_column} in (${6:values})\n);$0")
               snip(ft, "unpivot",
                  "select * from ${1:table}\nunpivot (\n\t${2:value_column}\n\tfor ${3:category_column} in (${4:columns})\n);$0")
               snip(ft, "xmlquery", "xmlquery('${1:xpath}' passing ${2:xml_column} returning content)$0")
               snip(ft, "jsontable",
                  "json_table(${1:json_column}, '${2:$}' columns (\n\t${3:column_name} ${4:datatype} path '${5:$.path}'\n))$0")

               snip(ft, "hint", "/*+ ${1:parallel(4)} */")
               snip(ft, "withhint",
                  "select /*+ ${1:index(t idx_name)} */ ${2:*}\nfrom ${3:table} t\nwhere ${4:condition};$0")
               snip(ft, "explain",
                  "explain plan for\n${1:select * from table};\nselect * from table(dbms_xplan.display);$0")
               snip(ft, "bulkcollect",
                  "declare\n\ttype t_tab is table of ${1:table}%rowtype;\n\tl_data t_tab;\nbegin\n\tselect * bulk collect into l_data\n\tfrom ${1:table}\n\twhere ${2:condition};\n\t\n\tforall i in 1..l_data.count\n\t\t${3:-- dml operation}\nend;\n/$0")

               snip(ft, "package",
                  "create or replace package ${1:package_name} as\n\t${2:-- declarations}\nend ${1:package_name};\n/$0")
               snip(ft, "packagebody",
                  "create or replace package body ${1:package_name} as\n\t${2:-- implementations}\nend ${1:package_name};\n/$0")
               snip(ft, "proc",
                  "create or replace procedure ${1:proc_name}(${2:p_param varchar2}) is\n\t${3:-- variables}\nbegin\n\t${4:-- logic}\nexception\n\twhen others then\n\t\tca_log_pak.log_error('${1:proc_name}', sqlerrm);\n\t\traise;\nend ${1:proc_name};\n/$0")
               snip(ft, "func",
                  "create or replace function ${1:func_name}(${2:p_param varchar2}) return ${3:varchar2} is\n\tl_result ${3:varchar2}(4000);\nbegin\n\t${4:-- logic}\n\treturn l_result;\nexception\n\twhen others then\n\t\tca_log_pak.log_error('${1:func_name}', sqlerrm);\n\t\treturn null;\nend ${1:func_name};\n/$0")
               snip(ft, "trigger",
                  "create or replace trigger ${1:trigger_name}\n${2:before} ${3:insert or update or delete} on ${4:table_name}\nfor each row\nbegin\n\t${5:-- logic}\nend ${1:trigger_name};\n/$0")

               snip(ft, "cursor",
                  "cursor ${1:c_name} is\n\tselect ${2:*}\n\tfrom ${3:table}\n\twhere ${4:condition};$0")
               snip(ft, "forcursor", "for rec in ${1:cursor_name} loop\n\t${2:-- process rec}\nend loop;$0")
               snip(ft, "forselect",
                  "for rec in (select ${1:*} from ${2:table} where ${3:condition}) loop\n\t${4:-- process rec}\nend loop;$0")
               snip(ft, "forloop", "for i in 1..${1:10} loop\n\t${2:-- process}\nend loop;$0")
               snip(ft, "while", "while ${1:condition} loop\n\t${2:-- process}\nend loop;$0")
               snip(ft, "loop", "loop\n\t${1:-- process}\n\texit when ${2:condition};\nend loop;$0")
               snip(ft, "if",
                  "if ${1:condition} then\n\t${2:-- true branch}\nelsif ${3:condition} then\n\t${4:-- elsif branch}\nelse\n\t${5:-- else branch}\nend if;$0")

               snip(ft, "exception",
                  "exception\n\twhen no_data_found then\n\t\t${1:-- handle}\n\twhen too_many_rows then\n\t\t${2:-- handle}\n\twhen others then\n\t\tca_log_pak.log_error('${3:source}', sqlerrm);\n\t\traise;$0")
               snip(ft, "pragma", "pragma ${1:autonomous_transaction};$0")
               snip(ft, "autonomoustx",
                  "declare\n\tpragma autonomous_transaction;\nbegin\n\t${1:-- independent transaction}\n\tcommit;\nend;\n/$0")
               snip(ft, "raiseapp", "raise_application_error(-20${1:001}, '${2:Error message}');$0")

               snip(ft, "type", "type ${1:t_name} is ${2:table of varchar2(100)};$0")
               snip(ft, "tabletype", "type ${1:t_tab} is table of ${2:table_name}%rowtype;$0")
               snip(ft, "recordtype",
                  "type ${1:t_rec} is record (\n\t${2:field1} ${3:varchar2(100)},\n\t${4:field2} ${5:number}\n);$0")
               snip(ft, "pipelined",
                  "create or replace function ${1:func_name} return ${2:t_table} pipelined is\nbegin\n\tfor rec in (select * from ${3:table}) loop\n\t\tpipe row(rec);\n\tend loop;\n\treturn;\nend;\n/$0")

               snip(ft, "dynamicsql", "execute immediate '${1:sql statement}' ${2:using ${3:bind_vars}};$0")
               snip(ft, "open", "open ${1:cursor_name};$0")
               snip(ft, "fetch", "fetch ${1:cursor_name} into ${2:variable};$0")
               snip(ft, "close", "close ${1:cursor_name};$0")
               snip(ft, "returning", "returning ${1:column} into ${2:variable}$0")
               snip(ft, "forall", "forall i in 1..${1:collection}.count\n\t${2:-- dml statement}$0")
               snip(ft, "indices", "forall i in indices of ${1:collection}\n\t${2:-- dml statement}$0")
               snip(ft, "values", "forall i in values of ${1:collection}\n\t${2:-- dml statement}$0")

               snip(ft, "dbmsoutput", "dbms_output.put_line('${1:message}: ' || ${2:variable});$0")
               snip(ft, "log_pak", "ca_log_pak.log_warning('${1:source}', '${2:message}: ' || ${3:variable});$0")

               snip(ft, "lbchangeset",
                  "--liquibase formatted sql\n\n--changeset ${1:author}:${2:id} runOnChange:${3:true} failOnError:true\n${4:-- sql statement}\n--rollback not supported$0")
               snip(ft, "lbchangesetplsql",
                  "--liquibase formatted sql\n\n--changeset ${1:author}:${2:id} runOnChange:true failOnError:true endDelimiter:\"/\"\ncreate or replace ${3:procedure} ${4:name} as\nbegin\n\t${5:-- logic}\nend;\n/\n--rollback not supported$0")
               snip(ft, "lbprecondition",
                  "--preconditions onFail:MARK_RAN\n--precondition-sql-check expectedResult:${1:0} select count(*) from ${2:table} where ${3:condition};$0")
               snip(ft, "lbcontext",
                  "--changeset ${1:author}:${2:id} context:${3:dev,test}\n${4:-- sql for specific contexts}$0")
               snip(ft, "lbsqlfile",
                  "--changeset ${1:author}:${2:id} runOnChange:true\n--sqlFile path:${3:file.sql} splitStatements:${4:true} stripComments:${5:true} endDelimiter:${6:;}$0")
               snip(ft, "lbformatted", "--liquibase formatted sql\n\n${1:-- changesets}$0")

               snip(ft, "changelog",
                  "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<databaseChangeLog\n\txmlns=\"http://www.liquibase.org/xml/ns/dbchangelog\"\n\txmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"\n\txsi:schemaLocation=\"http://www.liquibase.org/xml/ns/dbchangelog\n\thttp://www.liquibase.org/xml/ns/dbchangelog/dbchangelog-3.6.xsd\">\n\n\t<includeAll path=\"${1:tables}/\" relativeToChangelogFile=\"true\" />\n\n</databaseChangeLog>$0")
               snip(ft, "changelogfull",
                  "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<databaseChangeLog\n\txmlns=\"http://www.liquibase.org/xml/ns/dbchangelog\"\n\txmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"\n\txsi:schemaLocation=\"http://www.liquibase.org/xml/ns/dbchangelog\n\thttp://www.liquibase.org/xml/ns/dbchangelog/dbchangelog-3.6.xsd\">\n\n\t<property name=\"product\" value=\"${1:'tdssuite'}\"/>\n\t<property name=\"version\" value=\"${2:'{{ssm-app-version}}'}\"/>\n\t<property name=\"client_name\" value=\"${3:'{{ssm-instance-name}}'}\"/>\n\n\t<includeAll path=\"${4:tables}/\" relativeToChangelogFile=\"true\" />\n\n</databaseChangeLog>$0")
               snip(ft, "changeset",
                  "<include file=\"${1:path/to/file.sql}\" relativeToChangelogFile=\"true\" />$0")
               snip(ft, "include", "<include file=\"${1:path/to/file.xml}\" relativeToChangelogFile=\"${2:true}\"/>$0")
               snip(ft, "includeall", "<includeAll path=\"${1:directory}/\" relativeToChangelogFile=\"${2:true}\"/>$0")
               snip(ft, "property", "<property name=\"${1:product}\" value=\"${2:'{{ssm-app-version}}'}\"/>$0")
               snip(ft, "precondition",
                  "<preConditions onFail=\"MARK_RAN\">\n\t<sqlCheck expectedResult=\"0\">\n\t\tselect count(*) from user_tables where lower(table_name) = '${1:table_name}'\n\t</sqlCheck>\n</preConditions>$0")
               snip(ft, "createtable",
                  "<changeSet id=\"${1:id}\" author=\"${2:author}\">\n\t<createTable tableName=\"${3:table_name}\">\n\t\t<column name=\"id\" type=\"int\">\n\t\t\t<constraints primaryKey=\"true\"/>\n\t\t</column>\n\t\t<column name=\"${4:name}\" type=\"varchar(${5:50})\">\n\t\t\t<constraints nullable=\"false\"/>\n\t\t</column>\n\t</createTable>\n</changeSet>$0")
               snip(ft, "addcolumn",
                  "<changeSet id=\"${1:id}\" author=\"${2:author}\">\n\t<addColumn tableName=\"${3:table_name}\">\n\t\t<column name=\"${4:column_name}\" type=\"${5:varchar(50)}\"/>\n\t</addColumn>\n</changeSet>$0")
            end

            local js_filetypes = { "javascript", "typescript", "javascriptreact", "typescriptreact" }
            for _, ft in ipairs(js_filetypes) do
               snip(ft, "clwo", "console.warn('', {\n\t'${1}': ${1}\n});$0")
               snip(ft, "clw", "console.warn('${1}', ${1})")
               snip(ft, "clg", "console.log('${1}');$0")
               snip(ft, "clo", "console.log('${1}Obj', ${2}Obj);$0")
               snip(ft, "ccl", "console.clear();$0")
               snip(ft, "cer", "console.error('${1}');$0")
               snip(ft, "ctr", "console.trace();$0")
               snip(ft, "clt", "console.table('${1}');$0")
               snip(ft, "cin", "console.info('${1}');$0")
               snip(ft, "cco", "console.count('${1}');$0")
               snip(ft, "db", "debugger;$0")
               snip(ft, "deb", "debugger;$0")
               snip(ft, "if", "if (${1}) {\n\t$0\n}")
               snip(ft, "log_warn", "console.warn('${1}', ${1})")
               snip(ft, "log_warn_obj", "console.warn({\n\t'${1}': ${1}\n});$0")
               snip(ft, "tryc", "try {\n\t${1}\n} catch (error) {\n\tconsole.error('An error occurred:', error);\n}")
               snip(ft, "arrow", "(${1:params}) => ${2:expression}$0")
               snip(ft, "arrowf", "(${1:params}) => {\n\t${2}\n}$0")
               snip(ft, "asyncf", "async (${1:params}) => {\n\t${2}\n}$0")
               snip(ft, "promise", "new Promise((resolve, reject) => {\n\t${1}\n})$0")
               snip(ft, "then", ".then((${1:result}) => {\n\t${2}\n})$0")
               snip(ft, "catch", ".catch((${1:error}) => {\n\t${2}\n})$0")
               snip(ft, "finally", ".finally(() => {\n\t${1}\n})$0")
               snip(ft, "await", "await ${1:promise}$0")
               snip(ft, "import", "import ${1:{ ${2} }} from '${3:module}';$0")
               snip(ft, "export", "export ${1:const} ${2:name} = ${3:value};$0")
               snip(ft, "exportd", "export default ${1:expression};$0")
               snip(ft, "const", "const ${1:name} = ${2:value};$0")
               snip(ft, "let", "let ${1:name} = ${2:value};$0")
               snip(ft, "for", "for (let ${1:i} = 0; ${1:i} < ${2:array}.length; ${1:i}++) {\n\t${3}\n}$0")
               snip(ft, "forof", "for (const ${1:item} of ${2:array}) {\n\t${3}\n}$0")
               snip(ft, "forin", "for (const ${1:key} in ${2:object}) {\n\t${3}\n}$0")
               snip(ft, "foreach", "${1:array}.forEach((${2:item}) => {\n\t${3}\n});$0")
               snip(ft, "map", "${1:array}.map((${2:item}) => ${3:item})$0")
               snip(ft, "filter", "${1:array}.filter((${2:item}) => ${3:condition})$0")
               snip(ft, "reduce", "${1:array}.reduce((${2:acc}, ${3:item}) => {\n\t${4}\n}, ${5:initial})$0")
               snip(ft, "find", "${1:array}.find((${2:item}) => ${3:condition})$0")
               snip(ft, "includes", "${1:array}.includes(${2:value})$0")
               snip(ft, "spread", "...${1:array}$0")
               snip(ft, "dest", "const { ${1:prop} } = ${2:object};$0")
               snip(ft, "desta", "const [${1:first}, ${2:second}] = ${3:array};$0")
               snip(ft, "ternary", "${1:condition} ? ${2:true} : ${3:false}$0")
               snip(ft, "switch",
                  "switch (${1:expression}) {\n\tcase ${2:value}:\n\t\t${3}\n\t\tbreak;\n\tdefault:\n\t\t${4}\n}$0")
               snip(ft, "class", "class ${1:ClassName} {\n\tconstructor(${2:params}) {\n\t\t${3}\n\t}\n}$0")
               snip(ft, "extends",
                  "class ${1:ChildClass} extends ${2:ParentClass} {\n\tconstructor(${3:params}) {\n\t\tsuper(${4:params});\n\t\t${5}\n\t}\n}$0")
               snip(ft, "get", "get ${1:property}() {\n\treturn this.${2:_property};\n}$0")
               snip(ft, "set", "set ${1:property}(${2:value}) {\n\tthis.${3:_property} = ${2:value};\n}$0")
               snip(ft, "static", "static ${1:method}(${2:params}) {\n\t${3}\n}$0")
               snip(ft, "fetch",
                  "fetch('${1:url}')\n\t.then(response => response.json())\n\t.then(data => {\n\t\t${2}\n\t})\n\t.catch(error => console.error('Error:', error));$0")
               snip(ft, "fetchasync",
                  "try {\n\tconst response = await fetch('${1:url}');\n\tconst data = await response.json();\n\t${2}\n} catch (error) {\n\tconsole.error('Error:', error);\n}$0")
               snip(ft, "timeout", "setTimeout(() => {\n\t${1}\n}, ${2:1000});$0")
               snip(ft, "interval", "setInterval(() => {\n\t${1}\n}, ${2:1000});$0")
            end

            local ts_filetypes = { "typescript", "typescriptreact" }
            for _, ft in ipairs(ts_filetypes) do
               snip(ft, "interface", "interface ${1:Name} {\n\t${2:property}: ${3:type};\n}$0")
               snip(ft, "type", "type ${1:Name} = ${2:Type};$0")
               snip(ft, "enum", "enum ${1:Name} {\n\t${2:Value} = ${3:'value'},\n}$0")
               snip(ft, "generic", "<${1:T}>$0")
               snip(ft, "readonly", "readonly ${1:property}: ${2:type};$0")
               snip(ft, "partial", "Partial<${1:Type}>$0")
               snip(ft, "required", "Required<${1:Type}>$0")
               snip(ft, "record", "Record<${1:Key}, ${2:Value}>$0")
               snip(ft, "pick", "Pick<${1:Type}, ${2:'property'}>$0")
               snip(ft, "omit", "Omit<${1:Type}, ${2:'property'}>$0")
               snip(ft, "union", "${1:Type1} | ${2:Type2}$0")
               snip(ft, "intersection", "${1:Type1} & ${2:Type2}$0")
               snip(ft, "as", "as ${1:Type}$0")
               snip(ft, "is", "${1:param} is ${2:Type}$0")
               snip(ft, "namespace", "namespace ${1:Name} {\n\t${2}\n}$0")
               snip(ft, "declare", "declare ${1:const} ${2:name}: ${3:type};$0")
            end

            local react_filetypes = { "javascriptreact", "typescriptreact" }
            for _, ft in ipairs(react_filetypes) do
               snip(ft, "rfc",
                  "import React from 'react';\n\nconst ${1:ComponentName} = (${2:props}) => {\n\treturn (\n\t\t<div>\n\t\t\t${3}\n\t\t</div>\n\t);\n};\n\nexport default ${1:ComponentName};$0")
               snip(ft, "rfce",
                  "import React from 'react';\n\nconst ${1:ComponentName} = (${2:props}) => {\n\treturn (\n\t\t<div>\n\t\t\t${3}\n\t\t</div>\n\t);\n};\n\nexport default ${1:ComponentName};$0")
               snip(ft, "rafce",
                  "import React from 'react';\n\nconst ${1:ComponentName} = (${2:props}) => {\n\treturn (\n\t\t<>\n\t\t\t${3}\n\t\t</>\n\t);\n};\n\nexport default ${1:ComponentName};$0")
               snip(ft, "useState", "const [${1:state}, set${2:State}] = useState(${3:initialValue});$0")
               snip(ft, "useEffect", "useEffect(() => {\n\t${1}\n}, [${2:dependencies}]);$0")
               snip(ft, "useContext", "const ${1:context} = useContext(${2:Context});$0")
               snip(ft, "useReducer",
                  "const [${1:state}, ${2:dispatch}] = useReducer(${3:reducer}, ${4:initialState});$0")
               snip(ft, "useCallback", "const ${1:callback} = useCallback(() => {\n\t${2}\n}, [${3:dependencies}]);$0")
               snip(ft, "useMemo", "const ${1:memoized} = useMemo(() => ${2:computation}, [${3:dependencies}]);$0")
               snip(ft, "useRef", "const ${1:ref} = useRef(${2:initialValue});$0")
               snip(ft, "useLayoutEffect", "useLayoutEffect(() => {\n\t${1}\n}, [${2:dependencies}]);$0")
               snip(ft, "useImperativeHandle",
                  "useImperativeHandle(${1:ref}, () => ({\n\t${2}\n}), [${3:dependencies}]);$0")
               snip(ft, "custom-hook", "const use${1:HookName} = (${2:params}) => {\n\t${3}\n\treturn ${4:value};\n};$0")
               snip(ft, "jsx", "<${1:Component} ${2:props}>\n\t${3}\n</${1:Component}>$0")
               snip(ft, "jsxe", "<${1:Component} ${2:props} />$0")
               snip(ft, "jsxf", "<>\n\t${1}\n</>$0")
               snip(ft, "props", "${1:prop}={${2:value}}$0")
               snip(ft, "propsobj", "{...${1:props}}$0")
               snip(ft, "map-jsx",
                  "{${1:array}.map((${2:item}, ${3:index}) => (\n\t<${4:Component} key={${3:index}}>\n\t\t${5}\n\t</${4:Component}>\n))}$0")
               snip(ft, "conditional", "{${1:condition} && (\n\t${2}\n)}$0")
               snip(ft, "ternary-jsx", "{${1:condition} ? (\n\t${2}\n) : (\n\t${3}\n)}$0")
            end

            snip("sh", "shebang", "#!/bin/sh\n$0")
            snip("python", "shebang", "#!/usr/bin/env python3\n\n$0")

            snip("lua", "shebang", "#!/usr/bin/lua\n\n$0")
            snip("lua", "req", "require('${1:Module-name}')\n$0")
            snip("lua", "func", "function(${1:Arguments})\n\t${2}\nend\n$0")
            snip("lua", "forp", "for ${1:k}, ${2:v} in pairs(${3:table}) do\n\t${4}\nend\n$0")
            snip("lua", "fori", "for ${1:k}, ${2:v} in ipairs(${3:table}) do\n\t${4}\nend\n$0")
            snip("lua", "if", "if ${1} then\n\t${2}\nend\n$0")
            snip("lua", "M", "local M = {}\n$0\nreturn M")

            local text_filetypes = { "markdown", "text" }
            for _, ft in ipairs(text_filetypes) do
               snip(ft, "img", "- ![${2:alt text}](${1:})$0")
               snip(ft, "image", "- ![${2:alt text}](${1:})$0")
               snip(ft, "link", "- [${2:link text}](${1:})$0")
               snip(ft, "url", "- [${2:link text}](${1:})$0")
               snip(ft, "codewrap", "```${1:Language}\n${2}\n```\n$0")
               snip(ft, "code", "```${2:Language}\n${1:}\n```\n$0")
               snip(ft, "notes", "### NOTES: $0")
               snip(ft, "todo-list",
                  "## TODO\n\n- [ ] ${1:First task}\n- [ ] ${2:Second task}\n- [ ] ${3:Third task}\n$0")
               snip(ft, "table",
                  "| ${1:Header 1} | ${2:Header 2} | ${3:Header 3} |\n|------------|------------|------------|\n| ${4:Row 1} | ${5:Data} | ${6:Data} |\n| ${7:Row 2} | ${8:Data} | ${9:Data} |\n$0")
               snip(ft, "details",
                  "<details>\n<summary>${1:Click to expand}</summary>\n\n${2:Hidden content}\n\n</details>\n$0")
               snip(ft, "badge",
                  "![${1:Badge Name}](https://img.shields.io/badge/${2:label}-${3:message}-${4:color})")
               snip(ft, "h1", "# ${1:Title}\n$0")
               snip(ft, "h2", "## ${1:Section}\n$0")
               snip(ft, "h3", "### ${1:Subsection}\n$0")
               snip(ft, "h4", "#### ${1:Subsubsection}\n$0")
               snip(ft, "h5", "##### ${1:Paragraph}\n$0")
               snip(ft, "bold", "**${1:bold text}**$0")
               snip(ft, "italic", "*${1:italic text}*$0")
               snip(ft, "strike", "~~${1:strikethrough}~~$0")
               snip(ft, "quote", "> ${1:quote}\n$0")
               snip(ft, "codeblock", "```${1:language}\n${2:code}\n```\n$0")
               snip(ft, "inlinecode", "`${1:code}`$0")
               snip(ft, "hr", "---\n$0")
               snip(ft, "br", "<br>\n$0")
               snip(ft, "checkbox", "- [ ] ${1:task}$0")
               snip(ft, "checkedbox", "- [x] ${1:completed task}$0")
               snip(ft, "list", "- ${1:item 1}\n- ${2:item 2}\n- ${3:item 3}\n$0")
               snip(ft, "numlist", "1. ${1:item 1}\n2. ${2:item 2}\n3. ${3:item 3}\n$0")
               snip(ft, "ref", "[${1:text}][${2:ref}]\n\n[${2:ref}]: ${3:url}$0")
               snip(ft, "footnote", "${1:text}[^${2:1}]\n\n[^${2:1}]: ${3:footnote text}$0")
               snip(ft, "comment", "<!-- ${1:comment} -->$0")
               snip(ft, "yaml", "---\n${1:key}: ${2:value}\n---\n$0")
               snip(ft, "toc",
                  "## Table of Contents\n\n- [${1:Section 1}](#${2:section-1})\n- [${3:Section 2}](#${4:section-2})\n$0")
               snip(ft, "mermaid", "```mermaid\n${1:graph TD}\n    ${2:A[Start] --> B[End]}\n```\n$0")
               snip(ft, "diagram", "```${1:plantuml}\n${2:@startuml\n\n@enduml}\n```\n$0")
               snip(ft, "math", "$$${1:formula}$$")
               snip(ft, "inlinemath", "$${1:formula}$")
               snip(ft, "alert", "> [!${1:NOTE}]\n> ${2:content}$0")
               snip(ft, "warning", "> [!WARNING]\n> ${2:content}$0")
               snip(ft, "important", "> [!IMPORTANT]\n> ${2:content}$0")
               snip(ft, "tip", "> [!TIP]\n> ${2:content}$0")
               snip(ft, "caution", "> [!CAUTION]\n> ${2:content}$0")

               for trigger, language in pairs(code_languages) do
                  snip(ft, trigger, "```" .. language .. "\n${1}\n```$0")
               end
            end

            local css_filetypes = { "css", "scss" }
            for _, ft in ipairs(css_filetypes) do
               snip(ft, "flex", "display: flex;\njustify-content: ${1:center};\nalign-items: ${2:center};$0")
               snip(ft, "flexcol",
                  "display: flex;\nflex-direction: column;\njustify-content: ${1:center};\nalign-items: ${2:center};$0")
               snip(ft, "grid", "display: grid;\ngrid-template-columns: ${1:repeat(3, 1fr)};\ngrid-gap: ${2:1rem};$0")
               snip(ft, "gridarea",
                  "display: grid;\ngrid-template-areas:\n\t\"${1:header header}\"\n\t\"${2:sidebar main}\"\n\t\"${3:footer footer}\";$0")
               snip(ft, "center", "display: flex;\njustify-content: center;\nalign-items: center;$0")
               snip(ft, "transition", "transition: ${1:all} ${2:0.3s} ${3:ease};$0")
               snip(ft, "animation", "animation: ${1:name} ${2:1s} ${3:ease} ${4:infinite};$0")
               snip(ft, "keyframes", "@keyframes ${1:name} {\n\t0% {\n\t\t${2}\n\t}\n\t100% {\n\t\t${3}\n\t}\n}$0")
               snip(ft, "media", "@media (${1:max-width: 768px}) {\n\t${2}\n}$0")
               snip(ft, "var", "var(--${1:variable-name})$0")
               snip(ft, "root", ":root {\n\t--${1:variable-name}: ${2:value};\n}$0")
               snip(ft, "class", ".${1:class-name} {\n\t${2}\n}$0")
               snip(ft, "id", "#${1:id-name} {\n\t${2}\n}$0")
               snip(ft, "hover", "&:hover {\n\t${1}\n}$0")
               snip(ft, "before", "&::before {\n\tcontent: \"${1}\";\n\t${2}\n}$0")
               snip(ft, "after", "&::after {\n\tcontent: \"${1}\";\n\t${2}\n}$0")
               snip(ft, "shadow", "box-shadow: ${1:0} ${2:2px} ${3:4px} ${4:rgba(0, 0, 0, 0.1)};$0")
               snip(ft, "gradient", "background: linear-gradient(${1:to right}, ${2:#000}, ${3:#fff});$0")
               snip(ft, "border", "border: ${1:1px} ${2:solid} ${3:#000};$0")
               snip(ft, "padding", "padding: ${1:1rem};$0")
               snip(ft, "margin", "margin: ${1:1rem};$0")
               snip(ft, "font", "font-family: ${1:'Segoe UI'}, ${2:Tahoma}, ${3:sans-serif};$0")
            end

            snip("go", "iferr", "if err != nil {\n\t${1:return err}\n}$0")
            snip("go", "errnil", "if err != nil {\n\treturn ${1:nil, }err\n}$0")
            snip("go", "errlog", "if err != nil {\n\tlog.Printf(\"Error: %v\", err)\n\treturn err\n}$0")
            snip("go", "errwrap", "if err != nil {\n\treturn fmt.Errorf(\"${1:failed to}: %w\", err)\n}$0")
            snip("go", "func", "func ${1:name}(${2:params}) ${3:returnType} {\n\t${4}\n}$0")
            snip("go", "method",
               "func (${1:receiver} *${2:Type}) ${3:MethodName}(${4:params}) ${5:returnType} {\n\t${6}\n}$0")
            snip("go", "httphandler", "func ${1:handler}(w http.ResponseWriter, r *http.Request) {\n\t${2}\n}$0")
            snip("go", "struct", "type ${1:Name} struct {\n\t${2:Field} ${3:Type}\n}$0")
            snip("go", "interface", "type ${1:Name} interface {\n\t${2:Method}(${3:params}) ${4:returnType}\n}$0")
            snip("go", "test", "func Test${1:Name}(t *testing.T) {\n\t${2}\n}$0")
            snip("go", "bench",
               "func Benchmark${1:Name}(b *testing.B) {\n\tfor i := 0; i < b.N; i++ {\n\t\t${2}\n\t}\n}$0")
            snip("go", "goroutine", "go func() {\n\t${1}\n}()$0")
            snip("go", "wg",
               "var wg sync.WaitGroup\nwg.Add(${1:1})\ngo func() {\n\tdefer wg.Done()\n\t${2}\n}()\nwg.Wait()$0")
            snip("go", "defer", "defer ${1:func()}$0")
            snip("go", "ctx", "ctx := context.${1:Background()}$0")
            snip("go", "ctxtimeout",
               "ctx, cancel := context.WithTimeout(context.Background(), ${1:5*time.Second})\ndefer cancel()$0")
            snip("go", "switch", "switch ${1:expression} {\ncase ${2:value}:\n\t${3}\ndefault:\n\t${4}\n}$0")
            snip("go", "select", "select {\ncase ${1:msg} := <-${2:channel}:\n\t${3}\ndefault:\n\t${4}\n}$0")
            snip("go", "for", "for ${1:i} := ${2:0}; ${1:i} < ${3:10}; ${1:i}++ {\n\t${4}\n}$0")
            snip("go", "forrange", "for ${1:i}, ${2:v} := range ${3:slice} {\n\t${4}\n}$0")
            snip("go", "forindex", "for i := 0; i < len(${1:slice}); i++ {\n\t${2:// use slice[i]}\n}$0")
            snip("go", "forever", "for {\n\t${1:// infinite loop}\n\tif ${2:condition} {\n\t\tbreak\n\t}\n}$0")
            snip("go", "make", "${1:slice} := make([]${2:Type}, ${3:0})")
            snip("go", "map", "${1:m} := make(map[${2:string}]${3:interface{}})")
            snip("go", "chan", "${1:ch} := make(chan ${2:Type}${3:, bufferSize})")
            snip("go", "init", "func init() {\n\t${1}\n}$0")
            snip("go", "main", "func main() {\n\t${1}\n}$0")
            snip("go", "package", "package ${1:main}\n\n$0")
            snip("go", "import", "import (\n\t\"${1:fmt}\"\n)$0")

            snip("python", "main", "if __name__ == \"__main__\":\n\t${1:main()}$0")
            snip("python", "class", "class ${1:ClassName}:\n\tdef __init__(self${2:, args}):\n\t\t${3:pass}\n$0")
            snip("python", "def", "def ${1:function_name}(${2:args}):\n\t\"\"\"${3:Description}\"\"\"\n\t${4:pass}$0")
            snip("python", "method",
               "def ${1:method_name}(self${2:, args}):\n\t\"\"\"${3:Description}\"\"\"\n\t${4:pass}$0")
            snip("python", "tryf",
               "try:\n\t${1:pass}\nexcept ${2:Exception} as e:\n\t${3:print(f\"Error: {e}\")}\nfinally:\n\t${4:pass}$0")
            snip("python", "trye", "try:\n\t${1:pass}\nexcept ${2:Exception} as e:\n\t${3:print(f\"Error: {e}\")}$0")
            snip("python", "with", "with ${1:open(${2:file})} as ${3:f}:\n\t${4:pass}$0")
            snip("python", "decorator",
               "def ${1:decorator}(func):\n\tdef wrapper(*args, **kwargs):\n\t\t${2:# Before function}\n\t\tresult = func(*args, **kwargs)\n\t\t${3:# After function}\n\t\treturn result\n\treturn wrapper$0")
            snip("python", "async", "async def ${1:function_name}(${2:args}):\n\t${3:pass}$0")
            snip("python", "await", "await ${1:async_function()}$0")
            snip("python", "dataclass",
               "from dataclasses import dataclass\n\n@dataclass\nclass ${1:ClassName}:\n\t${2:field}: ${3:Type}$0")
            snip("python", "lambda", "lambda ${1:x}: ${2:x}$0")
            snip("python", "listcomp", "[${1:expr} for ${2:item} in ${3:iterable}${4: if condition}]$0")
            snip("python", "dictcomp", "{${1:key}: ${2:value} for ${3:item} in ${4:iterable}${5: if condition}}$0")
            snip("python", "setcomp", "{${1:expr} for ${2:item} in ${3:iterable}${4: if condition}}$0")
            snip("python", "genexp", "(${1:expr} for ${2:item} in ${3:iterable}${4: if condition})$0")
            snip("python", "property",
               "@property\ndef ${1:name}(self):\n\treturn self._${1}\n\n@${1}.setter\ndef ${1}(self, value):\n\tself._${1} = value$0")
            snip("python", "staticmethod", "@staticmethod\ndef ${1:method_name}(${2:args}):\n\t${3:pass}$0")
            snip("python", "classmethod", "@classmethod\ndef ${1:method_name}(cls${2:, args}):\n\t${3:pass}$0")
            snip("python", "super", "super().__init__(${1:args})$0")
            snip("python", "import", "import ${1:module}$0")
            snip("python", "from", "from ${1:module} import ${2:name}$0")
            snip("python", "if", "if ${1:condition}:\n\t${2:pass}$0")
            snip("python", "elif", "elif ${1:condition}:\n\t${2:pass}$0")
            snip("python", "else", "else:\n\t${1:pass}$0")
            snip("python", "for", "for ${1:item} in ${2:iterable}:\n\t${3:pass}$0")
            snip("python", "fori", "for i in range(${1:10}):\n\t${2:pass}$0")
            snip("python", "forelse", "for ${1:item} in ${2:iterable}:\n\t${3:pass}\nelse:\n\t${4:# no break}$0")
            snip("python", "forenum", "for ${1:index}, ${2:value} in enumerate(${3:iterable}):\n\t${4:pass}$0")
            snip("python", "forzip", "for ${1:a}, ${2:b} in zip(${3:list1}, ${4:list2}):\n\t${5:pass}$0")
            snip("python", "while", "while ${1:condition}:\n\t${2:pass}$0")
            snip("python", "match",
               "match ${1:expression}:\n\tcase ${2:pattern}:\n\t\t${3:pass}\n\tcase _:\n\t\t${4:pass}$0")
            snip("python", "raise", "raise ${1:Exception}(${2:\"Error message\"})$0")
            snip("python", "assert", "assert ${1:condition}, ${2:\"Assertion message\"}$0")
            snip("python", "print", "print(f\"${1:text}: {${2:variable}}\")$0")
            snip("python", "input", "${1:var} = input(\"${2:Enter value}: \")$0")
            snip("python", "open", "with open(\"${1:filename}\", \"${2:r}\") as f:\n\t${3:content = f.read()}$0")
            snip("python", "test", "def test_${1:name}():\n\t\"\"\"Test ${2:description}\"\"\"\n\t${3:assert True}$0")
            snip("python", "unittest",
               "import unittest\n\nclass Test${1:ClassName}(unittest.TestCase):\n\tdef test_${2:method}(self):\n\t\t${3:self.assertEqual(True, True)}$0")
            snip("python", "pytest", "import pytest\n\ndef test_${1:name}():\n\t${2:assert True}$0")

            snip("rust", "fn", "fn ${1:function_name}(${2:params}) ${3:-> ReturnType }{\n\t${4:todo!()}\n}$0")
            snip("rust", "pfn", "pub fn ${1:function_name}(${2:params}) ${3:-> ReturnType }{\n\t${4:todo!()}\n}$0")
            snip("rust", "impl", "impl ${1:Type} {\n\t${2}\n}$0")
            snip("rust", "implfor", "impl ${1:Trait} for ${2:Type} {\n\t${3}\n}$0")
            snip("rust", "struct", "struct ${1:Name} {\n\t${2:field}: ${3:Type},\n}$0")
            snip("rust", "pstruct", "pub struct ${1:Name} {\n\t${2:field}: ${3:Type},\n}$0")
            snip("rust", "enum", "enum ${1:Name} {\n\t${2:Variant},\n}$0")
            snip("rust", "penum", "pub enum ${1:Name} {\n\t${2:Variant},\n}$0")
            snip("rust", "trait", "trait ${1:Name} {\n\t${2:fn method(&self);}\n}$0")
            snip("rust", "match", "match ${1:expression} {\n\t${2:pattern} => ${3:value},\n\t_ => ${4:default},\n}$0")
            snip("rust", "matchopt", "match ${1:option} {\n\tSome(${2:value}) => ${3},\n\tNone => ${4},\n}$0")
            snip("rust", "matchres", "match ${1:result} {\n\tOk(${2:value}) => ${3},\n\tErr(${4:err}) => ${5},\n}$0")
            snip("rust", "iflet", "if let ${1:Some(value)} = ${2:expression} {\n\t${3}\n}$0")
            snip("rust", "whilelet", "while let ${1:Some(value)} = ${2:expression} {\n\t${3}\n}$0")
            snip("rust", "result", "Result<${1:T}, ${2:E}>$0")
            snip("rust", "option", "Option<${1:T}>$0")
            snip("rust", "vec", "vec![${1}]$0")
            snip("rust", "vecnew", "Vec::new()$0")
            snip("rust", "unwrap", ".unwrap()$0")
            snip("rust", "expect", ".expect(\"${1:error message}\")$0")
            snip("rust", "ok", "Ok(${1:value})$0")
            snip("rust", "err", "Err(${1:error})$0")
            snip("rust", "some", "Some(${1:value})$0")
            snip("rust", "none", "None$0")
            snip("rust", "derive", "#[derive(${1:Debug, Clone})]$0")
            snip("rust", "test", "#[test]\nfn test_${1:name}() {\n\t${2:assert_eq!(1, 1);}\n}$0")
            snip("rust", "cfg", "#[cfg(${1:test})]$0")
            snip("rust", "allow", "#[allow(${1:dead_code})]$0")
            snip("rust", "main", "fn main() {\n\t${1:println!(\"Hello, world!\");}\n}$0")
            snip("rust", "println", "println!(\"${1}\");$0")
            snip("rust", "print", "print!(\"${1}\");$0")
            snip("rust", "dbg", "dbg!(${1});$0")
            snip("rust", "todo", "todo!(${1})$0")
            snip("rust", "unimplemented", "unimplemented!(${1})$0")
            snip("rust", "panic", "panic!(\"${1}\")$0")
            snip("rust", "assert", "assert!(${1:condition});$0")
            snip("rust", "asserteq", "assert_eq!(${1:left}, ${2:right});$0")
            snip("rust", "assertne", "assert_ne!(${1:left}, ${2:right});$0")
            snip("rust", "use", "use ${1:std::collections::HashMap};$0")
            snip("rust", "mod", "mod ${1:module_name};$0")
            snip("rust", "pubmod", "pub mod ${1:module_name};$0")
            snip("rust", "loop", "loop {\n\t${1:break;}\n}$0")
            snip("rust", "for", "for ${1:item} in ${2:iterator} {\n\t${3}\n}$0")
            snip("rust", "forindexed", "for (${1:index}, ${2:item}) in ${3:iter}.enumerate() {\n\t${4}\n}$0")
            snip("rust", "forzip", "for (${1:a}, ${2:b}) in ${3:iter1}.zip(${4:iter2}) {\n\t${5}\n}$0")
            snip("rust", "forchunks", "for ${1:chunk} in ${2:slice}.chunks(${3:size}) {\n\t${4}\n}$0")
            snip("rust", "while", "while ${1:condition} {\n\t${2}\n}$0")
            snip("rust", "if", "if ${1:condition} {\n\t${2}\n}$0")
            snip("rust", "else", "else {\n\t${1}\n}$0")
            snip("rust", "elseif", "else if ${1:condition} {\n\t${2}\n}$0")
            snip("rust", "closure", "|${1:params}| ${2:expression}$0")
            snip("rust", "async", "async fn ${1:function_name}(${2:params}) ${3:-> ReturnType }{\n\t${4:todo!()}\n}$0")
            snip("rust", "await", ".await$0")

            snip("html", "html5",
               "<!DOCTYPE html>\n<html lang=\"${1:en}\">\n<head>\n\t<meta charset=\"UTF-8\">\n\t<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n\t<title>${2:Document}</title>\n</head>\n<body>\n\t${3}\n</body>\n</html>$0")
            snip("html", "div", "<div${1: class=\"${2}\"}>\n\t${3}\n</div>$0")
            snip("html", "span", "<span${1: class=\"${2}\"}>${3}</span>$0")
            snip("html", "a", "<a href=\"${1:#}\"${2: target=\"_blank\"}>${3:Link text}</a>$0")
            snip("html", "img", "<img src=\"${1}\" alt=\"${2}\"${3: width=\"${4}\" height=\"${5}\"}>$0")
            snip("html", "form", "<form${1: action=\"${2}\" method=\"${3:POST}\"}>\n\t${4}\n</form>$0")
            snip("html", "input", "<input type=\"${1:text}\" name=\"${2}\" id=\"${3}\"${4: placeholder=\"${5}\"}>$0")
            snip("html", "button", "<button type=\"${1:button}\"${2: class=\"${3}\"}>${4:Click me}</button>$0")
            snip("html", "ul", "<ul>\n\t<li>${1}</li>\n\t<li>${2}</li>\n\t<li>${3}</li>\n</ul>$0")
            snip("html", "ol", "<ol>\n\t<li>${1}</li>\n\t<li>${2}</li>\n\t<li>${3}</li>\n</ol>$0")
            snip("html", "table",
               "<table>\n\t<thead>\n\t\t<tr>\n\t\t\t<th>${1:Header 1}</th>\n\t\t\t<th>${2:Header 2}</th>\n\t\t</tr>\n\t</thead>\n\t<tbody>\n\t\t<tr>\n\t\t\t<td>${3:Data 1}</td>\n\t\t\t<td>${4:Data 2}</td>\n\t\t</tr>\n\t</tbody>\n</table>$0")
            snip("html", "section", "<section${1: class=\"${2}\"}>\n\t${3}\n</section>$0")
            snip("html", "article", "<article${1: class=\"${2}\"}>\n\t${3}\n</article>$0")
            snip("html", "header", "<header${1: class=\"${2}\"}>\n\t${3}\n</header>$0")
            snip("html", "footer", "<footer${1: class=\"${2}\"}>\n\t${3}\n</footer>$0")
            snip("html", "nav", "<nav${1: class=\"${2}\"}>\n\t${3}\n</nav>$0")
            snip("html", "main", "<main${1: class=\"${2}\"}>\n\t${3}\n</main>$0")
            snip("html", "script", "<script${1: src=\"${2}\"}>\n\t${3}\n</script>$0")
            snip("html", "link", "<link rel=\"${1:stylesheet}\" href=\"${2:style.css}\">$0")
            snip("html", "meta", "<meta ${1:name}=\"${2}\" content=\"${3}\">$0")

            snip("dockerfile", "from", "FROM ${1:node:18-alpine}$0")
            snip("dockerfile", "run", "RUN ${1:npm install}$0")
            snip("dockerfile", "cmd", "CMD [\"${1:npm}\", \"${2:start}\"]$0")
            snip("dockerfile", "entrypoint", "ENTRYPOINT [\"${1:node}\"]$0")
            snip("dockerfile", "copy", "COPY ${1:.} ${2:/app}$0")
            snip("dockerfile", "add", "ADD ${1:source} ${2:destination}$0")
            snip("dockerfile", "workdir", "WORKDIR ${1:/app}$0")
            snip("dockerfile", "expose", "EXPOSE ${1:3000}$0")
            snip("dockerfile", "env", "ENV ${1:NODE_ENV}=${2:production}$0")
            snip("dockerfile", "arg", "ARG ${1:VERSION}=${2:latest}$0")
            snip("dockerfile", "user", "USER ${1:node}$0")
            snip("dockerfile", "volume", "VOLUME [\"${1:/data}\"]$0")
            snip("dockerfile", "label", "LABEL ${1:maintainer}=\"${2:email@example.com}\"$0")
            snip("dockerfile", "healthcheck",
               "HEALTHCHECK --interval=${1:30s} --timeout=${2:3s} --start-period=${3:5s} --retries=${4:3} \\\n\tCMD ${5:curl -f http://localhost/ || exit 1}$0")
            snip("dockerfile", "multistage",
               "# Build stage\nFROM ${1:node:18-alpine} AS builder\nWORKDIR /app\nCOPY package*.json ./\nRUN npm ci\nCOPY . .\nRUN npm run build\n\n# Production stage\nFROM ${1:node:18-alpine}\nWORKDIR /app\nCOPY --from=builder /app/dist ./dist\nCOPY --from=builder /app/node_modules ./node_modules\nEXPOSE ${2:3000}\nCMD [\"node\", \"dist/index.js\"]$0")

            snip("yaml", "key", "${1:key}: ${2:value}$0")
            snip("yaml", "list", "${1:items}:\n  - ${2:item1}\n  - ${3:item2}$0")
            snip("yaml", "dict", "${1:object}:\n  ${2:key}: ${3:value}$0")
            snip("yaml", "anchor", "${1:name}: &${2:anchor}\n  ${3:key}: ${4:value}$0")
            snip("yaml", "alias", "<<: *${1:anchor}$0")
            snip("yaml", "compose",
               "version: '3.8'\n\nservices:\n  ${1:app}:\n    image: ${2:node:18-alpine}\n    ports:\n      - \"${3:3000}:${4:3000}\"\n    environment:\n      - ${5:NODE_ENV=production}\n    volumes:\n      - ${6:./}:${7:/app}\n    command: ${8:npm start}$0")
            snip("yaml", "service",
               "${1:service}:\n  image: ${2:image:tag}\n  container_name: ${3:name}\n  ports:\n    - \"${4:8080}:${5:80}\"\n  environment:\n    - ${6:ENV_VAR}=${7:value}\n  volumes:\n    - ${8:./data}:${9:/data}\n  networks:\n    - ${10:network}\n  restart: ${11:unless-stopped}$0")
            snip("yaml", "network", "networks:\n  ${1:network}:\n    driver: ${2:bridge}$0")
            snip("yaml", "volume", "volumes:\n  ${1:data}:\n    driver: ${2:local}$0")
            snip("yaml", "env", "environment:\n  - ${1:KEY}=${2:value}$0")
            snip("yaml", "ports", "ports:\n  - \"${1:8080}:${2:80}\"$0")
            snip("yaml", "depends", "depends_on:\n  - ${1:service}$0")
            snip("yaml", "healthcheck",
               "healthcheck:\n  test: [\"CMD\", \"${1:curl}\", \"-f\", \"${2:http://localhost/}\"]\n  interval: ${3:30s}\n  timeout: ${4:10s}\n  retries: ${5:3}\n  start_period: ${6:40s}$0")
            snip("yaml", "deploy",
               "deploy:\n  replicas: ${1:3}\n  resources:\n    limits:\n      cpus: '${2:0.5}'\n      memory: ${3:512M}\n    reservations:\n      cpus: '${4:0.25}'\n      memory: ${5:256M}$0")

            snip("json", "obj", "{\n\t\"${1:key}\": \"${2:value}\"\n}$0")
            snip("json", "arr", "[\n\t${1}\n]$0")
            snip("json", "kv", "\"${1:key}\": \"${2:value}\"$0")
         end

         local filetype_cache = {}

         local function get_buf_snips()
            local ft = vim.bo.filetype

            if not filetype_cache[ft] then
               local snips = {}

               if registry.all then
                  for _, snippet in pairs(registry.all) do
                     table.insert(snips, snippet)
                  end
               end

               if ft and registry[ft] then
                  for _, snippet in pairs(registry[ft]) do
                     table.insert(snips, snippet)
                  end
               end

               filetype_cache[ft] = snips
            end

            return filetype_cache[ft]
         end

         local function register_cmp_source()
            local cmp_source = {}
            local cache = setmetatable({}, { __mode = 'k' })

            function cmp_source.complete(_, _, callback)
               local bufnr = vim.api.nvim_get_current_buf()
               if not cache[bufnr] then
                  local completion_items = vim.tbl_map(function(s)
                     local body = s.body
                     if type(body) == "function" then
                        body = body()
                     end

                     local item = {
                        word = s.trigger,
                        label = s.trigger,
                        kind = vim.lsp.protocol.CompletionItemKind.Snippet,
                        insertText = body,
                        insertTextFormat = vim.lsp.protocol.InsertTextFormat.Snippet,
                     }
                     return item
                  end, get_buf_snips())

                  cache[bufnr] = completion_items
               end

               callback(cache[bufnr])
            end

            local augroup = vim.api.nvim_create_augroup('SnippetCache', { clear = true })
            vim.api.nvim_create_autocmd({ "BufEnter", "FileType", "BufDelete" }, {
               group = augroup,
               callback = function(args)
                  cache[args.buf] = nil

                  if args.event == "FileType" then
                     filetype_cache[vim.bo[args.buf].filetype] = nil
                  end
               end,
            })

            require("cmp").register_source("native_snippets", cmp_source)
         end

         vim.opt.completeopt = { "menu", "menuone", "noselect" }

         local cmp = require("cmp")
         local select_opts = { behavior = cmp.SelectBehavior.Insert }

         setup_snippets()

         register_cmp_source()

         cmp.setup({
            snippet = {
               expand = function(args)
                  vim.snippet.expand(args.body)
               end,
            },
            sources = {
               { name = "native_snippets", keyword_length = 1, priority = 1000,    max_item_count = 5,  group_index = 1 },
               { name = "nvim_lsp",        keyword_length = 3, priority = 900,     max_item_count = 10, group_index = 2 },
               { name = "path",            priority = 800,     max_item_count = 5, group_index = 3 },
               { name = "buffer",          keyword_length = 3, priority = 700,     max_item_count = 5,  group_index = 3 },
            },
            sorting = {
               priority_weight = 2,
               comparators = {
                  cmp.config.compare.exact,
                  cmp.config.compare.score,
                  cmp.config.compare.recently_used,
                  cmp.config.compare.locality,
                  cmp.config.compare.kind,
                  cmp.config.compare.sort_text,
                  cmp.config.compare.length,
                  cmp.config.compare.order,
               },
            },
            performance = {
               max_view_entries = 15,
               debounce = 30,
               throttle = 15,
               fetching_timeout = 200,
               filtering_context_budget = 3,
               confirm_resolve_timeout = 80,
               async_budget = 1,
            },
            window = {
               documentation = cmp.config.window.bordered(),
            },
            formatting = {
               fields = { "menu", "abbr", "kind" },
               format = function(entry, item)
                  local menu_icon = {
                     nvim_lsp = "",
                     native_snippets = "",
                     buffer = "",
                     path = "",
                  }
                  item.menu = menu_icon[entry.source.name]
                  return item
               end,
            },
            mapping = {
               ["<Up>"] = cmp.mapping.select_prev_item(select_opts),
               ["<Down>"] = cmp.mapping.select_next_item(select_opts),
               ["<C-k>"] = cmp.mapping.select_prev_item(select_opts),
               ["<C-j>"] = cmp.mapping.select_next_item(select_opts),
               ["<C-p>"] = cmp.mapping.select_prev_item(select_opts),
               ["<C-n>"] = cmp.mapping.select_next_item(select_opts),
               ["<C-u>"] = cmp.mapping.scroll_docs(-4),
               ["<C-f>"] = cmp.mapping.scroll_docs(4),
               ["<C-e>"] = cmp.mapping.abort(),
               ["<C-y>"] = cmp.mapping.confirm({ select = true }),
               ["<CR>"] = cmp.mapping.confirm({ select = false }),
               ["<C-Space>"] = cmp.mapping.complete(),

               ["<C-d>"] = cmp.mapping(function(fallback)
                  if vim.snippet.active({ direction = 1 }) then
                     vim.snippet.jump(1)
                  else
                     fallback()
                  end
               end, { "i", "s" }),

               ["<C-b>"] = cmp.mapping(function(fallback)
                  if vim.snippet.active({ direction = -1 }) then
                     vim.snippet.jump(-1)
                  else
                     fallback()
                  end
               end, { "i", "s" }),

               ["<Tab>"] = cmp.mapping(function(fallback)
                  local col = vim.fn.col(".") - 1
                  if cmp.visible() then
                     cmp.select_next_item(select_opts)
                  elseif vim.snippet.active({ direction = 1 }) then
                     vim.snippet.jump(1)
                  elseif col == 0 or vim.fn.getline("."):sub(col, col):match("%s") then
                     fallback()
                  else
                     cmp.complete()
                  end
               end, { "i", "s" }),

               ["<S-Tab>"] = cmp.mapping(function(fallback)
                  if cmp.visible() then
                     cmp.select_prev_item(select_opts)
                  elseif vim.snippet.active({ direction = -1 }) then
                     vim.snippet.jump(-1)
                  else
                     fallback()
                  end
               end, { "i", "s" }),
            },
         })

         vim.keymap.set({ "i" }, "<C-s>e", function()
            local current_word = vim.fn.expand('<cword>')
            local filetype = vim.bo.filetype
            local snippet = nil

            if registry[filetype] and registry[filetype][current_word] then
               snippet = registry[filetype][current_word]
            elseif registry.all[current_word] then
               snippet = registry.all[current_word]
            end

            if snippet then
               vim.cmd('normal! diw')
               local body = snippet.body
               if type(body) == "function" then
                  body = body()
               end
               vim.snippet.expand(body)
            else
               vim.notify("No snippet found for: " .. current_word, vim.log.levels.INFO)
            end
         end, { desc = "Expand snippet under cursor", silent = true })

         vim.keymap.set({ "i", "s" }, "<C-s>;", function()
            if vim.snippet.active({ direction = 1 }) then
               vim.snippet.jump(1)
            end
         end, { desc = "Jump forward in snippet", silent = true })

         vim.keymap.set({ "i", "s" }, "<C-s>,", function()
            if vim.snippet.active({ direction = -1 }) then
               vim.snippet.jump(-1)
            end
         end, { desc = "Jump backward in snippet", silent = true })

         vim.keymap.set({ "i", "s" }, "<C-l>", function()
            if vim.snippet.active() then
               vim.snippet.stop()
            end
         end, { desc = "Stop snippet", silent = true })
      end,
   },
}

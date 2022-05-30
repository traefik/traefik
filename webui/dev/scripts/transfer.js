const fs = require('fs-extra')

const folder = process.argv[2]

async function execute () {
  try {
    await fs.emptyDir('./static')
    await fs.outputFile('./static/DONT-EDIT-FILES-IN-THIS-DIRECTORY.md', 'For more information see `webui/readme.md`')
    console.log('Deleted static folder contents!')
    await fs.copy(`./dist/${folder}`, './static', { overwrite: true })
    console.log('Installed new files in static folder!')
  } catch (err) {
    console.error(err)
  }
}

execute()

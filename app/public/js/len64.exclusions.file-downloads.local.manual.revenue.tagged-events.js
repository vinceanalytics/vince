!function(){"use strict";var u=window.location,o=window.document,p=o.currentScript,l=p.getAttribute("data-api")||new URL(p.src).origin+"/api/event";function s(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){try{if("true"===window.localStorage.plausible_ignore)return s("localStorage flag",t)}catch(e){}var r=p&&p.getAttribute("data-include"),a=p&&p.getAttribute("data-exclude");if("pageview"===e){r=!r||r.split(",").some(n),a=a&&a.split(",").some(n);if(!r||a)return s("exclusion rule",t)}function n(e){return u.pathname.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var r={},i=(r.n=e,r.u=t&&t.u?t.u:u.href,r.d=p.getAttribute("data-domain"),r.r=o.referrer||null,t&&t.meta&&(r.m=JSON.stringify(t.meta)),t&&t.props&&(r.p=t.props),t&&t.revenue&&(r.$=t.revenue),new XMLHttpRequest);i.open("POST",l,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(r)),i.onreadystatechange=function(){4===i.readyState&&t&&t.callback&&t.callback({status:i.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var r=0;r<t.length;r++)e.apply(this,t[r]);function c(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var f=1;function a(e){var t,r,a,n;if("auxclick"!==e.type||e.button===f)return t=function(e){for(;e&&(void 0===e.tagName||!c(e)||!e.href);)e=e.parentNode;return e}(e.target),r=t&&t.href&&t.href.split("?")[0],!function e(t,r){if(!t||b<r)return!1;if(h(t))return!0;return e(t.parentNode,r+1)}(t,0)&&(a=r)&&(n=a.split(".").pop(),d.some(function(e){return e===n}))?v(e,t,{name:"File Download",props:{url:r}}):void 0}function v(e,t,r){var a,n=!1;function i(){n||(n=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?((a={props:r.props}).revenue=r.revenue,plausible(r.name,a)):((a={props:r.props,callback:i}).revenue=r.revenue,plausible(r.name,a),setTimeout(i,5e3),e.preventDefault())}o.addEventListener("click",a),o.addEventListener("auxclick",a);var n=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],i=p.getAttribute("file-types"),m=p.getAttribute("add-file-types"),d=i&&i.split(",")||m&&m.split(",").concat(n)||n;function g(e){var e=h(e)?e:e&&e.parentNode,t={name:null,props:{},revenue:{}},r=e&&e.classList;if(r)for(var a=0;a<r.length;a++){var n,i,u=r.item(a),o=u.match(/plausible-event-(.+)(=|--)(.+)/),o=(o&&(n=o[1],i=o[3].replace(/\+/g," "),"name"==n.toLowerCase()?t.name=i:t.props[n]=i),u.match(/plausible-revenue-(.+)(=|--)(.+)/));o&&(n=o[1],i=o[3],t.revenue[n]=i)}return t}var b=3;function w(e){if("auxclick"!==e.type||e.button===f){for(var t,r,a,n,i=e.target,u=0;u<=b&&i;u++){if((a=i)&&a.tagName&&"form"===a.tagName.toLowerCase())return;c(i)&&(t=i),h(i)&&(r=i),i=i.parentNode}r&&(n=g(r),t?(n.props.url=t.href,v(e,t,n)):((e={}).props=n.props,e.revenue=n.revenue,plausible(n.name,e)))}}function h(e){var t=e&&e.classList;if(t)for(var r=0;r<t.length;r++)if(t.item(r).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}o.addEventListener("submit",function(e){var t,r=e.target,a=g(r);function n(){t||(t=!0,r.submit())}a.name&&(e.preventDefault(),t=!1,setTimeout(n,5e3),(e={props:a.props,callback:n}).revenue=a.revenue,plausible(a.name,e))}),o.addEventListener("click",w),o.addEventListener("auxclick",w)}();